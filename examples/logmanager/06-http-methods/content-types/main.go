package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "trace_id"

type User struct {
	ID       string            `json:"id" form:"id"`
	Name     string            `json:"name" form:"name"`
	Email    string            `json:"email" form:"email"`
	Age      int               `json:"age" form:"age"`
	Metadata map[string]string `json:"metadata,omitempty"`
}


func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-content-types"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders("Content-Type", "Content-Length"),
		logmanager.WithDebug(),
	)

	r := gin.Default()
	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	// Content-Type examples
	api := r.Group("/api/v1")
	{
		// JSON Content-Type
		api.POST("/json", handleJSON)

		// Form Data (multipart/form-data)
		api.POST("/form-data", handleFormData)

		// URL Encoded (application/x-www-form-urlencoded)
		api.POST("/urlencoded", handleURLEncoded)

		// Plain Text
		api.POST("/text", handlePlainText)

		// Binary Data
		api.POST("/binary", handleBinary)

		// XML Data
		api.POST("/xml", handleXML)

		// File Upload
		api.POST("/upload", handleFileUpload)

		// Multiple files upload
		api.POST("/upload-multiple", handleMultipleFileUpload)
	}


	// Raw data endpoints
	raw := r.Group("/raw")
	{
		raw.POST("/bytes", handleRawBytes)
		raw.POST("/stream", handleStream)
	}

	fmt.Println("HTTP Content-Types server running at http://localhost:8081")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /api/v1/json           - JSON payload")
	fmt.Println("  POST /api/v1/form-data      - multipart/form-data")
	fmt.Println("  POST /api/v1/urlencoded     - application/x-www-form-urlencoded")
	fmt.Println("  POST /api/v1/text           - text/plain")
	fmt.Println("  POST /api/v1/binary         - application/octet-stream")
	fmt.Println("  POST /api/v1/xml            - application/xml")
	fmt.Println("  POST /api/v1/upload         - file upload")
	fmt.Println("  POST /api/v1/upload-multiple - multiple files upload")
	fmt.Println("  POST /raw/bytes             - raw bytes")
	fmt.Println("  POST /raw/stream            - streaming data")

	log.Fatal(r.Run(":8081"))
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(traceIDKey, traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

// JSON Content-Type handler
func handleJSON(c *gin.Context) {
	logSegment(c, "handle-json", func() {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON",
				"details": err.Error(),
			})
			return
		}

		user.ID = uuid.NewString()
		c.JSON(http.StatusOK, gin.H{
			"message": "JSON processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"data": user,
		})
	})
}

// Form Data (multipart/form-data) handler
func handleFormData(c *gin.Context) {
	logSegment(c, "handle-form-data", func() {
		var user User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid form data",
				"details": err.Error(),
			})
			return
		}

		// Handle file if present
		file, header, err := c.Request.FormFile("avatar")
		var fileInfo map[string]interface{}
		if err == nil {
			defer file.Close()
			fileInfo = map[string]interface{}{
				"filename": header.Filename,
				"size":     header.Size,
				"contentType": header.Header.Get("Content-Type"),
			}
		}

		user.ID = uuid.NewString()
		c.JSON(http.StatusOK, gin.H{
			"message": "Form data processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"data": user,
			"file": fileInfo,
		})
	})
}

// URL Encoded handler
func handleURLEncoded(c *gin.Context) {
	logSegment(c, "handle-url-encoded", func() {
		var user User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid URL encoded data",
				"details": err.Error(),
			})
			return
		}

		user.ID = uuid.NewString()
		c.JSON(http.StatusOK, gin.H{
			"message": "URL encoded data processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"data": user,
		})
	})
}

// Plain Text handler
func handlePlainText(c *gin.Context) {
	logSegment(c, "handle-plain-text", func() {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read body",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Plain text processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"length": len(body),
			"content": string(body),
		})
	})
}

// Binary Data handler
func handleBinary(c *gin.Context) {
	logSegment(c, "handle-binary", func() {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read binary data",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Binary data processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"size": len(body),
			"md5": fmt.Sprintf("%x", body[:min(8, len(body))]), // First 8 bytes as hex
		})
	})
}

// XML handler
func handleXML(c *gin.Context) {
	logSegment(c, "handle-xml", func() {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read XML data",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "XML processed successfully",
			"contentType": c.GetHeader("Content-Type"),
			"length": len(body),
			"xml": string(body),
		})
	})
}

// File Upload handler
func handleFileUpload(c *gin.Context) {
	logSegment(c, "handle-file-upload", func() {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
				"details": err.Error(),
			})
			return
		}
		defer file.Close()

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read file",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			"filename": header.Filename,
			"size": header.Size,
			"contentType": header.Header.Get("Content-Type"),
			"actualSize": len(content),
		})
	})
}

// Multiple File Upload handler
func handleMultipleFileUpload(c *gin.Context) {
	logSegment(c, "handle-multiple-file-upload", func() {
		err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to parse multipart form",
				"details": err.Error(),
			})
			return
		}

		files := c.Request.MultipartForm.File["files"]
		var fileInfos []map[string]interface{}

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			content, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				continue
			}

			fileInfos = append(fileInfos, map[string]interface{}{
				"filename": fileHeader.Filename,
				"size": fileHeader.Size,
				"contentType": fileHeader.Header.Get("Content-Type"),
				"actualSize": len(content),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Multiple files uploaded successfully",
			"count": len(fileInfos),
			"files": fileInfos,
		})
	})
}


// Raw Bytes handler
func handleRawBytes(c *gin.Context) {
	logSegment(c, "handle-raw-bytes", func() {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read raw bytes",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Raw bytes processed successfully",
			"size": len(body),
			"contentType": c.GetHeader("Content-Type"),
			"sample": fmt.Sprintf("%x", body[:min(16, len(body))]), // First 16 bytes as hex
		})
	})
}

// Stream handler
func handleStream(c *gin.Context) {
	logSegment(c, "handle-stream", func() {
		buffer := bytes.NewBuffer(nil)
		bytesRead, err := io.Copy(buffer, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read stream",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Stream processed successfully",
			"bytesRead": bytesRead,
			"contentType": c.GetHeader("Content-Type"),
		})
	})
}

// Helper function for consistent segment logging
func logSegment(c *gin.Context, segmentName string, handler func()) {
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(c),
		logmanager.OtherSegment{
			Name: segmentName,
		},
	)
	defer txn.End()

	handler()
}

// Helper function (Go 1.21+ builtin, but included for compatibility)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Example usage functions for testing different content types
func createExampleRequests() {
	// JSON Example
	jsonExample := `{
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30,
		"metadata": {
			"department": "engineering",
			"level": "senior"
		}
	}`

	// URL Encoded Example
	urlEncodedExample := url.Values{
		"name":  {"John Doe"},
		"email": {"john@example.com"},
		"age":   {"30"},
	}.Encode()

	// Form Data Example
	formDataExample := func() *bytes.Buffer {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("name", "John Doe")
		w.WriteField("email", "john@example.com")
		w.WriteField("age", "30")
		w.Close()
		return &b
	}

	// Usage examples (not executed, just for documentation)
	_ = jsonExample
	_ = urlEncodedExample
	_ = formDataExample
}