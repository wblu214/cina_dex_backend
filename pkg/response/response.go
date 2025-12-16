package response

// Response defines a standard API response envelope.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success wraps a successful response with data.
func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Error wraps an error response with a custom code and message.
func Error(code int, msg string) Response {
	return Response{
		Code:    code,
		Message: msg,
	}
}
