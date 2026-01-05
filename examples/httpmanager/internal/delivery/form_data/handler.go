package form_data

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

// NewHandler creates a form data handler demonstrating custom error responses
func NewHandler() *httpmanager.UploadHandler {
	return httpmanager.NewUploadHandler(
		http.MethodPost,
		"./uploads",
		func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
			// Get form value using public helper
			name := httpmanager.GetFormValue(form, "name")

			// Validate name - return custom error if empty
			if strings.TrimSpace(name) == "" {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("name is required"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "FORM_001",
						Message: "Name field is required",
					},
				}
			}

			// Return success
			return &ContactFormResponse{
				Status:  "success",
				Message: "Form submitted",
				Data: ContactFormData{
					Name: name,
				},
			}, nil
		},
	)
}
