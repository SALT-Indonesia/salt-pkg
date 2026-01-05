package upload_validation

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

// NewHandler creates a new upload handler with validation and custom error responses
func NewHandler() *httpmanager.UploadHandler {
	return httpmanager.NewUploadHandler(
		http.MethodPost,
		"./uploads",
		func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
			// Validate required form field using public helper
			title := httpmanager.GetFormValue(form, "title")
			if strings.TrimSpace(title) == "" {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("title is required"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "UPLOAD_001",
						Message: "Title field is required",
					},
				}
			}

			// Validate that at least one file is uploaded
			if len(files) == 0 {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("no files uploaded"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "UPLOAD_002",
						Message: "At least one file must be uploaded",
					},
				}
			}

			// Collect file info
			var fileInfos []FileInfo
			for fieldName, uploadedFiles := range files {
				for _, file := range uploadedFiles {
					fileInfos = append(fileInfos, FileInfo{
						FieldName:   fieldName,
						Filename:    file.Filename,
						Size:        file.Size,
						ContentType: file.ContentType,
					})
				}
			}

			// Return success response
			return &UploadSuccessResponse{
				Status:  "success",
				Message: fmt.Sprintf("Uploaded %d file(s)", len(fileInfos)),
				Title:   title,
				Files:   fileInfos,
			}, nil
		},
	).WithMaxFileSize(10 << 20) // 10 MB max
}
