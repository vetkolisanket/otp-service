package service

import (
	"time"
	"github.com/go-redis/redis"
)

//OtpService - holds redis client instance
type OtpService struct {
	redisClient *redis.Client
}

//NewOtpService - Returns a new instance of OtpService
func NewOtpService(r *redis.Client) *OtpService {
	return &OtpService{r}
}

//PingRedis - Pings redis server
func (s *OtpService) PingRedis() (string, error) {
	return s.redisClient.Ping().Result()
}

//GetResultFromRedis - Returns the result stored in redis for the key passed in parameter
func (s *OtpService) GetResultFromRedis(key string) (string, error) {
	return s.redisClient.Get(key).Result()
}

//StoreResultToRedis - Stores the given data in redis corresponding to the key value for the specified duration
func (s *OtpService) StoreResultToRedis(key string, data interface{}, duration time.Duration) (string, error) {
	return s.redisClient.Set(key, data, duration).Result()
}