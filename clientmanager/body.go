package clientmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func stringify(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", val), "0"), ".")
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

// getMultipartFormBody creates multipart form body with files and values.
// This function handles both files (with custom content types) and string form fields.
func getMultipartFormBody(form MultipartForm) (*bytes.Buffer, string) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add files with custom content types
	for field, fileData := range form.Files {
		// Create custom MIME header with Content-Disposition and Content-Type
		h := make(map[string][]string)
		h["Content-Disposition"] = []string{
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, field, fileData.Filename),
		}
		h["Content-Type"] = []string{fileData.ContentType}

		part, _ := writer.CreatePart(h)

		_, _ = part.Write(fileData.Content)
	}

	// Add string form fields
	for field, value := range form.Values {
		_ = writer.WriteField(field, value)
	}

	_ = writer.Close()

	return body, writer.FormDataContentType()
}

func getFilesBody(files map[string]string, requestBody any) (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for field, path := range files {
		file, err := os.Open(filepath.Clean(path)) // #nosec G304 - file paths from user configuration
		if err != nil {
			return nil, "", err
		}
		defer func() {
			_ = file.Close()
		}()

		part, _ := writer.CreateFormFile(field, filepath.Base(path))
		_, _ = io.Copy(part, file)
	}
	if requestBody != nil { // add JSON fields if reqBody exists
		data, _ := json.Marshal(requestBody)
		var form map[string]any
		_ = json.Unmarshal(data, &form)
		for k, v := range form {
			_ = writer.WriteField(k, stringify(v))
		}
	}
	_ = writer.Close()
	return body, writer.FormDataContentType(), nil
}

func getFormURLEncodedBody(requestBody any) (*bytes.Buffer, string) {
	formData := url.Values{}
	data, _ := json.Marshal(requestBody)
	var form map[string]any
	_ = json.Unmarshal(data, &form)
	for k, v := range form {
		formData.Set(k, stringify(v))
	}
	return bytes.NewBufferString(formData.Encode()), "application/x-www-form-urlencoded"
}

func getJSONBody(requestBody any) *bytes.Buffer {
	if requestBody != nil {
		data, _ := json.Marshal(requestBody)
		return bytes.NewBuffer(data)
	}
	return nil
}
