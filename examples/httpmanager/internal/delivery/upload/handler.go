package upload

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"net/http"
)

type Handler struct {
}

func NewHandler() *httpmanager.UploadHandler {
	return httpmanager.NewUploadHandler(
		http.MethodPost, "./uploads", func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {

			for fieldName, uploadedFiles := range files {
				for _, file := range uploadedFiles {
					// Access file metadata
					fmt.Printf("Field: %s, File: %s, Size: %d, Type: %s, Saved at: %s\n",
						fieldName, file.Filename, file.Size, file.ContentType, file.SavedPath)
				}
			}

			// Access form values
			//name := ""
			//if values, ok := form["name"]; ok && len(values) > 0 {
			//	name = values[0]
			//}
			// Return a response
			return map[string]string{
				"status":  "success",
				"message": "Files uploaded successfully",
			}, nil
		},
	)
}
