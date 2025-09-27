package event

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"net/http"
)

type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type CreateEventResponse struct {
	ID int `json:"id"`
}

func NewHandler() *httpmanager.UploadHandler {
	return httpmanager.NewUploadHandler(
		http.MethodPost,
		"./uploads",
		func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
			// Extract form fields
			title := getFormValue(form, "title")
			description := getFormValue(form, "description")
			location := getFormValue(form, "location")
			startDate := getFormValue(form, "start_date")
			endDate := getFormValue(form, "end_date")

			// Process uploaded files (e.g., event poster/banner)
			var uploadedFileInfo []map[string]interface{}
			for fieldName, uploadedFiles := range files {
				for _, file := range uploadedFiles {
					uploadedFileInfo = append(uploadedFileInfo, map[string]interface{}{
						"field":        fieldName,
						"filename":     file.Filename,
						"size":         file.Size,
						"content_type": file.ContentType,
						"saved_path":   file.SavedPath,
					})
				}
			}

			// Return success response
			return map[string]interface{}{
				"status":  201,
				"message": "event created successfully",
				"data": CreateEventResponse{
					ID: 1,
				},
				"event_details": map[string]interface{}{
					"title":       title,
					"description": description,
					"location":    location,
					"start_date":  startDate,
					"end_date":    endDate,
					"attachments": uploadedFileInfo,
				},
			}, nil
		},
	)
}

func getFormValue(form map[string][]string, key string) string {
	if values, ok := form[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}