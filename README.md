# otp-service

Service which given a no. provides a one time password

1.Basic ping call. Done
2.Basic ping DB.
3.Get otp call - Creates an otp for the number provided along with otp token, stores it in redis with an expiration, and returns the otp and token in response. In progress.
4.Validate otp call.

## TODOs

*Separation of concern
*Validate otp
