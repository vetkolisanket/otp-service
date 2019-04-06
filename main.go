package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/vetkolisanket/otp-service/handlers"
	"github.com/vetkolisanket/otp-service/service"

	"github.com/go-redis/redis"
)

var (
	port      = flag.String("p", ":1234", "Port to run the otp service on")
	redisPort = flag.String("redisPort", "localhost:6379", "Redis port")
)

var redisClient *redis.Client

func main() {
	flag.Parse()

	r := NewRedisClient()

	s := service.NewOtpService(r)

	checkRedisStatus(s)

	h := handlers.NewHTTPHandler(s)

	server := &http.Server{
		Addr:              *port,
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
		Addr:     *redisPort,
		Password: "",
		DB:       0,
	})
}
