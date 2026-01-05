# File Uploads and Static Files

This document covers file upload handling and static file serving.

## File Upload Handling

The module provides support for handling file uploads via multipart form data:

```go
// Create an upload handler
uploadHandler := httpmanager.NewUploadHandler(
    http.MethodPost,
    "./uploads", // Directory where files will be saved
    func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
        // Process uploaded files
        for fieldName, uploadedFiles := range files {
            for _, file := range uploadedFiles {
                // Access file metadata
                fmt.Printf("Field: %s, File: %s, Size: %d, Type: %s, Saved at: %s\n",
                    fieldName, file.Filename, file.Size, file.ContentType, file.SavedPath)
            }
        }

        // Access form values
        name := ""
        if values, ok := form["name"]; ok && len(values) > 0 {
            name = values[0]
        }

        // Return a response
        return map[string]string{
            "status": "success",
            "message": "Files uploaded successfully",
        }, nil
    },
)

// Optionally set maximum file size (default is 10MB)
uploadHandler.WithMaxFileSize(20 << 20) // 20MB

// Add middleware if needed
uploadHandler.Use(authMiddleware)

// Register the handler
server.Handle("/upload", uploadHandler.WithMiddleware())
```

### UploadedFile Structure

The `UploadedFile` struct provides metadata about uploaded files:

```go
type UploadedFile struct {
	Filename    string // Original filename
	Size        int64  // File size in bytes
	ContentType string // MIME type
	SavedPath   string // Path where the file was saved
}
```

### Form Value Helper Functions

The module provides helper functions for accessing form values:

```go
// Get first value for a form field
name := httpmanager.GetFormValue(form, "name")

// Get all values for a form field (for multi-value fields)
tags := httpmanager.GetFormValues(form, "tags")
```

| Function | Description |
|----------|-------------|
| `GetFormValue(form, key)` | Returns the first value for the key, or empty string if not found |
| `GetFormValues(form, key)` | Returns all values for the key, or nil if not found |

### HTML Form Example

Here's an example HTML form that works with the upload handler:

```html
<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="text" name="name" placeholder="Your name">
    <input type="file" name="files" multiple>
    <input type="file" name="avatar">
    <button type="submit">Upload</button>
</form>
```

### Error Handling with ResponseError

The `UploadHandler` supports `ResponseError[T]` for custom error responses, allowing you to return structured JSON errors with appropriate HTTP status codes:

```go
// Define your error response format
type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Create an upload handler with validation and custom error responses
uploadHandler := httpmanager.NewUploadHandler(
    http.MethodPost,
    "./uploads",
    func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
        // Validate required form fields
        title := httpmanager.GetFormValue(form, "title")
        if title == "" {
            return nil, &httpmanager.ResponseError[ErrorResponse]{
                Err:        fmt.Errorf("title is required"),
                StatusCode: http.StatusBadRequest,
                Body: ErrorResponse{
                    Code:    "UPLOAD_VAL_001",
                    Message: "Title field is required",
                    Data:    nil,
                },
            }
        }

        // Validate that at least one file is uploaded
        if len(files) == 0 {
            return nil, &httpmanager.ResponseError[ErrorResponse]{
                Err:        fmt.Errorf("no files uploaded"),
                StatusCode: http.StatusBadRequest,
                Body: ErrorResponse{
                    Code:    "UPLOAD_VAL_002",
                    Message: "At least one file must be uploaded",
                    Data:    nil,
                },
            }
        }

        // Return success response
        return map[string]string{
            "status":  "success",
            "message": "Files uploaded successfully",
        }, nil
    },
)

server.Handle("/upload", uploadHandler)
```

### Upload Error Response Examples

**Missing required field (400):**
```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@image.jpg" \
  -F "description=test"
```

Response:
```json
{
  "code": "UPLOAD_VAL_001",
  "message": "Title field is required"
}
```

**No file uploaded (400):**
```bash
curl -X POST http://localhost:8080/upload \
  -F "title=My Document" \
  -F "description=test"
```

Response:
```json
{
  "code": "UPLOAD_VAL_002",
  "message": "At least one file must be uploaded"
}
```

## Static File Serving

The module provides support for serving static files, particularly images, with appropriate content types:

```go
// Create a static file handler
staticHandler := httpmanager.NewStaticHandler(
    http.MethodGet,
    "./static", // Directory containing static files
)

// Add middleware if needed
staticHandler.Use(cacheMiddleware)

// Register the handler
server.Handle("/images/", staticHandler.WithMiddleware())
```

The static handler automatically:
- Serves files from the specified directory
- Sets appropriate content type headers based on file extensions
- Applies cache control headers for better performance
- Prevents directory traversal attacks
- Handles common image formats (JPEG, PNG, GIF, SVG, WebP, etc.)

### Accessing Static Files

Once the static handler is registered, files can be accessed via URLs like:

```
https://yourserver.com/images/logo.png
https://yourserver.com/images/photos/vacation.jpg
```

The handler maps these URLs to files in the static directory:

```
./static/logo.png
./static/photos/vacation.jpg
```

### Supported Image Formats

The static handler supports the following image formats with appropriate content types:

| Extension   | Content Type  |
|-------------|---------------|
| .jpg, .jpeg | image/jpeg    |
| .png        | image/png     |
| .gif        | image/gif     |
| .svg        | image/svg+xml |
| .webp       | image/webp    |
| .ico        | image/x-icon  |
| .bmp        | image/bmp     |
| .tiff, .tif | image/tiff    |
