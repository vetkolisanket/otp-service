package models

import (
	"encoding/json"
)

//HTTPResponse - A standard format in which response is being sent
type HTTPResponse struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"responseData,omitempty"`
}

//GetOtpResponse - Response model for get otp call
type GetOtpResponse struct {
	Otp      int    `json:"otp,omitempty"`
	OtpToken string `json:"token,omitempty"`
}

//MarshalBinary - Used to store GetOtpResponse in redis
func (r *GetOtpResponse) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary - Used to retrieve GetOtpResponse from redis
func (r *GetOtpResponse) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &r)
}