package main

//ErrorMessage err message
type ErrorMessage string

//ResponseError response for errors
var ResponseError = Status{"error", "An Servererror occurred"}

//ResponseSuccess response for success but no data
var ResponseSuccess = Status{"success", ""}

const (
	//ServerError error from server
	ServerError ErrorMessage = "Server Error"
	//WrongLogType wrong user input
	WrongLogType ErrorMessage = "Wrong Logtype!"
	//WrongInputFormatError wrong user input
	WrongInputFormatError ErrorMessage = "Wrong inputFormat!"
	//InvalidTokenError token is not valid
	InvalidTokenError ErrorMessage = "Token not valid"
	//BatchSizeTooLarge batch is too large
	BatchSizeTooLarge ErrorMessage = "BatchSize soo large!"
	//WrongIntegerFormat integer is probably no integer
	WrongIntegerFormat ErrorMessage = "Number is string"
)

//Status a REST response status
type Status struct {
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
}
