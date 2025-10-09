package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "trace_id"


type User struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Age      int               `json:"age"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type StreamData struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Type      string      `json:"type"`
}

// Sample data
var users = map[string]*User{
	"1": {ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
	"2": {ID: "2", Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	"3": {ID: "3", Name: "Bob Johnson", Email: "bob@example.com", Age: 35},
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-advanced-features"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders("Content-Type", "Content-Length", "Content-Disposition"),
		logmanager.WithDebug(),
	)

	r := gin.Default()
	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	// Create uploads directory if it doesn't exist
	os.MkdirAll("./uploads", 0755)


	// File upload endpoints
	upload := r.Group("/upload")
	{
		// Single file upload
		upload.POST("/single", handleSingleFileUpload)

		// Multiple files upload
		upload.POST("/multiple", handleMultipleFilesUpload)

		// Chunked file upload
		upload.POST("/chunked", handleChunkedUpload)

		// Large file upload with progress
		upload.POST("/large", handleLargeFileUpload)

		// Base64 encoded file upload
		upload.POST("/base64", handleBase64Upload)
	}

	// Streaming endpoints
	streaming := r.Group("/stream")
	{
		// Server-Sent Events
		streaming.GET("/events", handleServerSentEvents)

		// Chunked response
		streaming.GET("/chunked", handleChunkedResponse)

		// Real-time data stream
		streaming.GET("/data", handleDataStream)

		// File download with streaming
		streaming.GET("/download/:filename", handleStreamingDownload)
	}

	// WebSocket-like endpoints (using HTTP)
	realtime := r.Group("/realtime")
	{
		// Long polling
		realtime.GET("/poll", handleLongPolling)

		// Webhook receiver
		realtime.POST("/webhook", handleWebhook)
	}

	// Advanced request/response patterns
	advanced := r.Group("/advanced")
	{
		// Conditional requests (ETags, If-Modified-Since)
		advanced.GET("/conditional", handleConditionalRequest)

		// CORS preflight
		advanced.OPTIONS("/cors", handleCORSPreflight)
		advanced.POST("/cors", handleCORSRequest)

		// Content compression
		advanced.GET("/compressed", handleCompressedResponse)

		// Partial content (Range requests)
		advanced.GET("/partial", handlePartialContent)
	}

	fmt.Println("Advanced HTTP Features server running at http://localhost:8083")
	fmt.Println("Available endpoints:")
	fmt.Println("File Upload:")
	fmt.Println("  POST /upload/single        - Single file upload")
	fmt.Println("  POST /upload/multiple      - Multiple files upload")
	fmt.Println("  POST /upload/chunked       - Chunked upload")
	fmt.Println("  POST /upload/large         - Large file with progress")
	fmt.Println("Streaming:")
	fmt.Println("  GET  /stream/events        - Server-Sent Events")
	fmt.Println("  GET  /stream/chunked       - Chunked response")
	fmt.Println("  GET  /stream/data          - Real-time data stream")

	log.Fatal(r.Run(":8083"))
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


// Single file upload
func handleSingleFileUpload(c *gin.Context) {
	logSegment(c, "handle-single-file-upload", func() {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
				"details": err.Error(),
			})
			return
		}
		defer file.Close()

		// Save file
		filename := uuid.NewString() + "_" + header.Filename
		filepath := filepath.Join("./uploads", filename)

		out, err := os.Create(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save file",
				"details": err.Error(),
			})
			return
		}
		defer out.Close()

		size, err := io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to write file",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			"filename": header.Filename,
			"saved_as": filename,
			"size": size,
			"content_type": header.Header.Get("Content-Type"),
		})
	})
}

// Multiple files upload
func handleMultipleFilesUpload(c *gin.Context) {
	logSegment(c, "handle-multiple-files-upload", func() {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid multipart form",
				"details": err.Error(),
			})
			return
		}

		files := form.File["files"]
		var uploadedFiles []map[string]interface{}

		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				continue
			}

			filename := uuid.NewString() + "_" + file.Filename
			filepath := filepath.Join("./uploads", filename)

			dst, err := os.Create(filepath)
			if err != nil {
				src.Close()
				continue
			}

			size, err := io.Copy(dst, src)
			src.Close()
			dst.Close()

			if err == nil {
				uploadedFiles = append(uploadedFiles, map[string]interface{}{
					"original_name": file.Filename,
					"saved_as":      filename,
					"size":          size,
					"content_type":  file.Header.Get("Content-Type"),
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Files uploaded successfully",
			"count":   len(uploadedFiles),
			"files":   uploadedFiles,
		})
	})
}

// Chunked file upload
func handleChunkedUpload(c *gin.Context) {
	logSegment(c, "handle-chunked-upload", func() {
		chunkNumber := c.PostForm("chunk")
		totalChunks := c.PostForm("total_chunks")
		filename := c.PostForm("filename")

		file, _, err := c.Request.FormFile("chunk_data")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No chunk data",
				"details": err.Error(),
			})
			return
		}
		defer file.Close()

		// Save chunk
		chunkFilename := fmt.Sprintf("%s.chunk.%s", filename, chunkNumber)
		chunkPath := filepath.Join("./uploads", chunkFilename)

		out, err := os.Create(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save chunk",
				"details": err.Error(),
			})
			return
		}
		defer out.Close()

		size, err := io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to write chunk",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Chunk uploaded successfully",
			"chunk": chunkNumber,
			"total_chunks": totalChunks,
			"size": size,
			"filename": filename,
		})
	})
}

// Large file upload with progress tracking
func handleLargeFileUpload(c *gin.Context) {
	logSegment(c, "handle-large-file-upload", func() {
		// Set max memory for parsing (32MB)
		c.Request.ParseMultipartForm(32 << 20)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
				"details": err.Error(),
			})
			return
		}
		defer file.Close()

		filename := uuid.NewString() + "_" + header.Filename
		filepath := filepath.Join("./uploads", filename)

		out, err := os.Create(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create file",
				"details": err.Error(),
			})
			return
		}
		defer out.Close()

		// Copy with progress (simplified)
		buffer := make([]byte, 32*1024) // 32KB buffer
		totalSize := int64(0)

		for {
			n, err := file.Read(buffer)
			if n > 0 {
				out.Write(buffer[:n])
				totalSize += int64(n)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to read file",
					"details": err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Large file uploaded successfully",
			"filename": header.Filename,
			"saved_as": filename,
			"size": totalSize,
			"expected_size": header.Size,
		})
	})
}

// Base64 encoded file upload
func handleBase64Upload(c *gin.Context) {
	logSegment(c, "handle-base64-upload", func() {
		var request struct {
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
			Data        string `json:"data"` // Base64 encoded
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON",
				"details": err.Error(),
			})
			return
		}

		// Decode base64 data (simplified - real implementation would use proper base64 decoding)
		data := []byte(request.Data) // In real scenario, use base64.StdEncoding.DecodeString()

		filename := uuid.NewString() + "_" + request.Filename
		filepath := filepath.Join("./uploads", filename)

		err := os.WriteFile(filepath, data, 0644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save file",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Base64 file uploaded successfully",
			"filename": request.Filename,
			"saved_as": filename,
			"size": len(data),
			"content_type": request.ContentType,
		})
	})
}

// Server-Sent Events
func handleServerSentEvents(c *gin.Context) {
	logSegment(c, "handle-server-sent-events", func() {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Send initial event
		c.SSEvent("message", map[string]interface{}{
			"id":        1,
			"timestamp": time.Now(),
			"data":      "Connection established",
		})

		// Send periodic events
		for i := 2; i <= 5; i++ {
			time.Sleep(1 * time.Second)
			c.SSEvent("message", map[string]interface{}{
				"id":        i,
				"timestamp": time.Now(),
				"data":      fmt.Sprintf("Event %d", i),
			})
			c.Writer.Flush()
		}

		c.SSEvent("close", "Stream ended")
	})
}

// Chunked response
func handleChunkedResponse(c *gin.Context) {
	logSegment(c, "handle-chunked-response", func() {
		c.Header("Transfer-Encoding", "chunked")
		c.Header("Content-Type", "application/json")

		// Start JSON array
		c.Writer.WriteString(`{"data":[`)
		c.Writer.Flush()

		// Send data in chunks
		for i := 0; i < 5; i++ {
			if i > 0 {
				c.Writer.WriteString(",")
			}

			chunk := map[string]interface{}{
				"chunk":     i + 1,
				"timestamp": time.Now(),
				"data":      fmt.Sprintf("Chunk %d data", i+1),
			}

			data, _ := json.Marshal(chunk)
			c.Writer.Write(data)
			c.Writer.Flush()

			time.Sleep(500 * time.Millisecond)
		}

		// End JSON array
		c.Writer.WriteString(`]}`)
	})
}

// Real-time data stream
func handleDataStream(c *gin.Context) {
	logSegment(c, "handle-data-stream", func() {
		c.Header("Content-Type", "application/x-ndjson") // Newline Delimited JSON

		for i := 0; i < 10; i++ {
			data := StreamData{
				ID:        uuid.NewString(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"value":   i,
					"message": fmt.Sprintf("Stream data %d", i),
				},
				Type: "data",
			}

			jsonData, _ := json.Marshal(data)
			c.Writer.Write(jsonData)
			c.Writer.WriteString("\n")
			c.Writer.Flush()

			time.Sleep(500 * time.Millisecond)
		}
	})
}

// Streaming file download
func handleStreamingDownload(c *gin.Context) {
	filename := c.Param("filename")

	logSegment(c, "handle-streaming-download", func() {
		filepath := filepath.Join("./uploads", filename)

		file, err := os.Open(filepath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get file info",
			})
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))

		// Stream file in chunks
		io.Copy(c.Writer, file)
	})
}

// Long polling
func handleLongPolling(c *gin.Context) {
	logSegment(c, "handle-long-polling", func() {
		timeout := 30 * time.Second
		start := time.Now()

		// Simulate waiting for data
		for time.Since(start) < timeout {
			// Check for new data (simplified)
			if time.Since(start) > 5*time.Second {
				c.JSON(http.StatusOK, gin.H{
					"message": "New data available",
					"timestamp": time.Now(),
					"waited": time.Since(start).Seconds(),
				})
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Timeout
		c.JSON(http.StatusNoContent, gin.H{
			"message": "No new data",
			"timeout": true,
		})
	})
}

// Webhook receiver
func handleWebhook(c *gin.Context) {
	logSegment(c, "handle-webhook", func() {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read webhook payload",
			})
			return
		}

		// Verify webhook signature (simplified)
		signature := c.GetHeader("X-Webhook-Signature")

		c.JSON(http.StatusOK, gin.H{
			"message": "Webhook received successfully",
			"signature": signature,
			"payload_size": len(body),
			"content_type": c.GetHeader("Content-Type"),
			"timestamp": time.Now(),
		})
	})
}

// Conditional requests
func handleConditionalRequest(c *gin.Context) {
	logSegment(c, "handle-conditional-request", func() {
		etag := `"123456789"`
		lastModified := "Wed, 21 Oct 2024 07:28:00 GMT"

		// Check If-None-Match header
		if c.GetHeader("If-None-Match") == etag {
			c.Status(http.StatusNotModified)
			return
		}

		// Check If-Modified-Since header
		if c.GetHeader("If-Modified-Since") == lastModified {
			c.Status(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
		c.Header("Last-Modified", lastModified)
		c.JSON(http.StatusOK, gin.H{
			"data": "Resource data",
			"etag": etag,
			"last_modified": lastModified,
		})
	})
}

// CORS preflight
func handleCORSPreflight(c *gin.Context) {
	logSegment(c, "handle-cors-preflight", func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
}

// CORS request
func handleCORSRequest(c *gin.Context) {
	logSegment(c, "handle-cors-request", func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS request handled successfully",
			"origin": c.GetHeader("Origin"),
		})
	})
}

// Compressed response
func handleCompressedResponse(c *gin.Context) {
	logSegment(c, "handle-compressed-response", func() {
		// Large JSON response that benefits from compression
		data := make([]map[string]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = map[string]interface{}{
				"id":          i,
				"name":        fmt.Sprintf("Item %d", i),
				"description": strings.Repeat("This is a long description. ", 10),
				"metadata": map[string]interface{}{
					"category": "test",
					"tags":     []string{"tag1", "tag2", "tag3"},
				},
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Large dataset (compression recommended)",
			"count":   len(data),
			"data":    data,
		})
	})
}

// Partial content (Range requests)
func handlePartialContent(c *gin.Context) {
	logSegment(c, "handle-partial-content", func() {
		rangeHeader := c.GetHeader("Range")

		content := strings.Repeat("This is line content. ", 1000)
		contentLength := len(content)

		if rangeHeader != "" {
			// Parse range header (simplified)
			if strings.HasPrefix(rangeHeader, "bytes=") {
				rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
				parts := strings.Split(rangeSpec, "-")

				start, _ := strconv.Atoi(parts[0])
				end := contentLength - 1
				if len(parts) > 1 && parts[1] != "" {
					end, _ = strconv.Atoi(parts[1])
				}

				if start < contentLength && end < contentLength && start <= end {
					c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
					c.Header("Accept-Ranges", "bytes")
					c.Data(http.StatusPartialContent, "text/plain", []byte(content[start:end+1]))
					return
				}
			}
		}

		c.Header("Accept-Ranges", "bytes")
		c.Data(http.StatusOK, "text/plain", []byte(content))
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