# otp-service

Service which given a no. provides a one time password

1.--Basic ping call.-- Done
2.--Basic ping DB.-- Done - pinging redis
3.--Get otp call-- Done - Creates an otp for the number provided along with otp token, stores it in redis with an expiration, and returns the otp and token in response.
4.--Validate otp call.-- Done - Checks if the number provided along with the otp and otp token is valid or not

## TODOs

*Separation of concern
*--Validate otp--
