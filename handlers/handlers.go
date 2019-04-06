package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vetkolisanket/otp-service/models"
	"github.com/vetkolisanket/otp-service/service"
)

const (
	otpValidTimeInMinutes = 5

	mobileNumber = "mobileNumber"
)

//HTTPHandler - handles routing of requests to appropriate functions
type HTTPHandler struct {
	service *service.OtpService
}

//NewHTTPHandler - Returns a new instance of HTTPHandler
func NewHTTPHandler(s *service.OtpService) *HTTPHandler {
	return &HTTPHandler{s}
}

//GetHandlerFunc - returns an http.HandlerFunc which is responsible for delegating requests to appropriate functions
func (h *HTTPHandler) GetHandlerFunc() http.HandlerFunc {
	mux := http.NewServeMux()

	mux.HandleFunc("/otp-service/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{"success":true}`)
	})
	mux.HandleFunc("/otp-service/v1/ping/redis", h.pingRedisHandler)
	mux.HandleFunc("/otp-service/v1/otp", h.getOtpHandler)
	mux.HandleFunc("/otp-service/v1/otp/validate", h.validateOtpHandler)
	return mux.ServeHTTP
}

func (h *HTTPHandler) pingRedisHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.service.PingRedis()

	if err != nil {
		log.Println(err)
		writeResponse(w, http.StatusInternalServerError, "Redis is down!", nil)
	} else {
		writeResponse(w, http.StatusOK, "Redis is up!", nil)
	}
}

func (h *HTTPHandler) getOtpHandler(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get(mobileNumber)

	if m == "" {
		writeResponse(w, http.StatusBadRequest, "Mobile number needed while requesting otp!", nil)
		return
	}

	//todo validate if its a valid mobile number

	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(9999)

	uuid, err := exec.Command("uuidgen").Output()

	if err != nil {
		log.Println("Error while generating otp token", err)
		writeResponse(w, http.StatusInternalServerError, "Something went wrong!", nil)
	}

	otpToken := strings.TrimSuffix(string(uuid), "\n")

	log.Printf("Mobile number %s, Otp %04d, UUID %s", m, otp, otpToken)

	//todo convert otp to string and store 4 digits even if you get lesser digits from rand
	response := models.GetOtpResponse{
		Otp:      otp,
		OtpToken: otpToken,
	}

	_, err = h.service.StoreResultToRedis(m, &response, otpValidTimeInMinutes*time.Minute)

	if err != nil {
		log.Println("Error while saving response in redis", err)
		writeResponse(w, http.StatusInternalServerError, "Something went wrong!", nil)
	}

	//todo - Don't send the otp in response, instead send the token in response and delegate otp to a message carrier
	writeResponse(w, 200, fmt.Sprintf("Your otp for mobile number %s is %04d. It is valid for next %d minutes", m, otp,
		otpValidTimeInMinutes), response)
}

func (h *HTTPHandler) validateOtpHandler(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get(mobileNumber)

	if m == "" {
		writeResponse(w, http.StatusBadRequest, "Mobile number needed while validating otp!", nil)
		return
	}

	o := r.URL.Query().Get("otp")

	if o == "" {
		writeResponse(w, http.StatusBadRequest, "OTP needed while validating otp!", nil)
		return
	}

	otp, err := strconv.Atoi(o)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, "Otp should be an int!", nil)
		return
	}

	token := r.URL.Query().Get("token")

	if token == "" {
		writeResponse(w, http.StatusBadRequest, "Invalid token!", nil)
		return
	}

	//todo validate if its a valid mobile number

	s, err := h.service.GetResultFromRedis(m)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, "Invalid/Expired otp!", nil)
		return
	}

	response := models.GetOtpResponse{}

	err = response.UnmarshalBinary([]byte(s))

	if err != nil {
		log.Println("Error while unmarshaling response", err)
		writeResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	if otp != response.Otp {
		writeResponse(w, http.StatusBadRequest, "Invalid/Expired otp!", nil)
		return
	}

	if token != response.OtpToken {
		writeResponse(w, http.StatusBadRequest, "Invalid/Expired otp!", nil)
		return
	}

	writeResponse(w, http.StatusOK, "Okay", nil)
}

func writeResponse(w http.ResponseWriter, code int, msg string, data interface{}) {
	status := false
	if code >= 200 && code < 300 {
		status = true
	}
	response := &models.HTTPResponse{Status: status, Code: code, Message: msg, Data: data}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

}
