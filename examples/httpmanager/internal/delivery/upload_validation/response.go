package upload_validation

// ErrorResponse is the custom error response structure
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UploadSuccessResponse is the response returned on successful file upload
type UploadSuccessResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Title   string     `json:"title"`
	Files   []FileInfo `json:"files"`
}

// FileInfo contains metadata about an uploaded file
type FileInfo struct {
	FieldName   string `json:"field_name"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}
