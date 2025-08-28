package validation

type CreateUserResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// Standard error response format
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}