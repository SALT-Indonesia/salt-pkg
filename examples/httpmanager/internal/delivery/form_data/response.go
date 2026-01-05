package form_data

// ErrorResponse is the custom error response structure
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ContactFormResponse is the success response
type ContactFormResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    ContactFormData `json:"data"`
}

// ContactFormData contains the submitted form data
type ContactFormData struct {
	Name string `json:"name"`
}
