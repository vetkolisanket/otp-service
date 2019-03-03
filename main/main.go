package main

import (
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

	otpValidTimeInMinutes = 5
)

var redisClient *redis.Client

type httpResponse struct {
	Status  bool        `json:"status,omitempty"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"responseData,omitempty"`
}

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

	RedisNewClient()

	mux.HandleFunc(serviceName+versionName+ping, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{"success":true}`)
	})
	mux.HandleFunc(serviceName+versionName+getOtp, getOtpHandler)
	return mux.ServeHTTP
}

func getOtpHandler(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("mobileNumber")

	if m == "" {
		writeResponse(w, http.StatusBadRequest, "Mobile number needed while requesting otp!", nil)
		return
	}

	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(9999)

	otpToken, err := exec.Command("uuidgen").Output()

	if err != nil {
		log.Println("Error while generating otp token", err)
	}

	log.Printf("Mobile number %s, Otp %04d, UUID %s", m, otp, otpToken)

	respone := getOtpResponse{
		Otp:      otp,
		OtpToken: string(otpToken[:]),
	}

	responseBytes, err := json.Marshal(respone)

	if err != nil {
		log.Println("Error while marshaling response!", err)
	}

	err = redisClient.Set(m, responseBytes, otpValidTimeInMinutes*time.Minute).Err()

	if err != nil {
		log.Println("Error while saving response in redis", err)
	}

	//todo - Don't send the otp in response, instead set the token in response and delegate otp to a message carrier
	writeResponse(w, 200, fmt.Sprintf("Your otp for mobile number %s is %04d. It is valid for next %d minutes", m, otp,
		otpValidTimeInMinutes), respone)
}

//RedisNewClient ...
func RedisNewClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping().Result()
	log.Println(pong, err)

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
	if code > 200 && code < 300 {
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
