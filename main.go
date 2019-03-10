package main

import (
	"github.com/vetkolisanket/otp-service/handlers"
	"github.com/vetkolisanket/otp-service/service"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)



var redisClient *redis.Client

func main() {
	r := NewRedisClient()

	s := service.NewOtpService(r)

	checkRedisStatus(s)

	h := handlers.NewHTTPHandler(s)

	server := &http.Server{
		Addr:              ":1234",
		Handler:           h.GetHandlerFunc(),
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      1 * time.Minute,
	}

	

	log.Println("Starting otp service...")

	log.Fatal(server.ListenAndServe())
}

func checkRedisStatus(s *service.OtpService) {
	_, err := s.PingRedis()

	if err != nil {
		log.Println("Redis is down!!!")
	} else {
		log.Println("Redis is up")
	}
}

//NewRedisClient ...
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}