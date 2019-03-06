package main

import (
	"strconv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"time"

	"github.com/go-redis/redis"
)

const (
	serviceName = "/otp-service"
	versionName = "/v1"

	ping        = "/ping"
	getOtp      = "/otp"
	validateOtp = "/otp/validate"

	otpValidTimeInMinutes = 5

	mobileNumber = "mobileNumber"
)

var redisClient *redis.Client

type httpResponse struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"responseData,omitempty"`
}

//getOtpResponse ...
type getOtpResponse struct {
	Otp      int    `json:"otp,omitempty"`
	OtpToken string `json:"token,omitempty"`
}

func main() {
	server := &http.Server{
		Addr:              ":1234",
		Handler:           NewHTTPHandler(),
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      1 * time.Minute,
	}

	log.Println("otp-service is up")

	log.Fatal(server.ListenAndServe())
}

//NewHTTPHandler provides handler for routing of api requests for otp-service
func NewHTTPHandler() http.HandlerFunc {
	mux := http.NewServeMux()

	InitRedisClient()

	mux.HandleFunc(serviceName+versionName+ping, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{"success":true}`)
	})
	mux.HandleFunc(serviceName+versionName+getOtp, getOtpHandler)
	mux.HandleFunc(serviceName+versionName+validateOtp, validateOtpHandler)
	return mux.ServeHTTP
}

func validateOtpHandler(w http.ResponseWriter, r *http.Request) {
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

	res := redisClient.Get(m)

	s, err := res.Result()

	if err != nil {
		writeResponse(w, http.StatusBadRequest, "Invalid/Expired otp!", nil)
		return
	}

	response := getOtpResponse{}

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

func getOtpHandler(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get(mobileNumber)

	if m == "" {
		writeResponse(w, http.StatusBadRequest, "Mobile number needed while requesting otp!", nil)
		return
	}

	//todo validate if its a valid mobile number

	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(9999)

	otpToken, err := exec.Command("uuidgen").Output()

	log.Printf("%T", otpToken)

	if err != nil {
		log.Println("Error while generating otp token", err)
	}

	log.Printf("Mobile number %s, Otp %04d, UUID %s", m, otp, otpToken)

	response := getOtpResponse{
		Otp:      otp,
		OtpToken: string(otpToken),
	}

	err = redisClient.Set(m, &response, otpValidTimeInMinutes*time.Minute).Err()

	if err != nil {
		log.Println("Error while saving response in redis", err)
	}

	//todo - Don't send the otp in response, instead set the token in response and delegate otp to a message carrier
	writeResponse(w, 200, fmt.Sprintf("Your otp for mobile number %s is %04d. It is valid for next %d minutes", m, otp,
		otpValidTimeInMinutes), response)
}

//MarshalBinary ...
func (r *getOtpResponse) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary -
func (r *getOtpResponse) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &r)
}

//InitRedisClient ...
func InitRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping().Result()
	log.Println(pong, err)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Redis is up!")
	}

	// err = redisClient.Set("key", "value", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// val, err := redisClient.Get("key").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("key", val)

	// val2, err := redisClient.Get("key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }
}

func writeResponse(w http.ResponseWriter, code int, msg string, data interface{}) {
	status := false
	if code >= 200 && code < 300 {
		status = true
	}
	response := &httpResponse{Status: status, Code: code, Message: msg, Data: data}
	dataBytes, err := json.Marshal(response)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(dataBytes)
}
