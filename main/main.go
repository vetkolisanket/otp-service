package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

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

	mux.HandleFunc("/otp-service/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{"success":true}`)
	})
	return mux.ServeHTTP
}

//RedisNewClient ...
func RedisNewClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}
