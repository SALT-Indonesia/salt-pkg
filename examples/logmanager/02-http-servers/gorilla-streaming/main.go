package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgorilla"
	"github.com/gorilla/mux"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-gorilla-streaming"),
		logmanager.WithTags("streaming", "sse"),
	)

	router := mux.NewRouter()
	router.Use(lmgorilla.Middleware(app))

	router.HandleFunc("/sse", handleSSE).Methods(http.MethodGet)
	router.HandleFunc("/chunked", handleChunked).Methods(http.MethodGet)
	router.HandleFunc("/ndjson", handleNDJSON).Methods(http.MethodGet)

	fmt.Println("Gorilla streaming server running at http://localhost:8080")
	fmt.Println("Test endpoints:")
	fmt.Println("  curl http://localhost:8080/sse      # Server-Sent Events")
	fmt.Println("  curl http://localhost:8080/chunked  # Chunked transfer encoding")
	fmt.Println("  curl http://localhost:8080/ndjson   # Newline-delimited JSON")

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

// handleSSE demonstrates Server-Sent Events streaming with http.Flusher
func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// This type assertion should now work with the fix
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Println("SSE client connected")

	// Stream 10 events with 1-second intervals
	for i := 0; i < 10; i++ {
		// Check if client disconnected
		select {
		case <-r.Context().Done():
			fmt.Println("SSE client disconnected")
			return
		default:
		}

		// Send event
		fmt.Fprintf(w, "id: %d\n", i)
		fmt.Fprintf(w, "data: Event %d at %s\n\n", i, time.Now().Format(time.RFC3339))

		// Flush to send immediately
		flusher.Flush()

		time.Sleep(1 * time.Second)
	}

	fmt.Println("SSE streaming completed")
}

// handleChunked demonstrates chunked transfer encoding with multiple flushes
func handleChunked(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Println("Chunked client connected")

	// Send 5 chunks with delays
	for i := 1; i <= 5; i++ {
		// Check if client disconnected
		select {
		case <-r.Context().Done():
			fmt.Println("Chunked client disconnected")
			return
		default:
		}

		// Write chunk
		chunk := fmt.Sprintf("Chunk %d: %s\n", i, time.Now().Format(time.RFC3339))
		fmt.Fprint(w, chunk)

		// Flush immediately
		flusher.Flush()

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Fprintln(w, "All chunks sent!")
	fmt.Println("Chunked streaming completed")
}

// handleNDJSON demonstrates newline-delimited JSON streaming
func handleNDJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Println("NDJSON client connected")

	encoder := json.NewEncoder(w)

	// Stream 8 JSON objects
	for i := 1; i <= 8; i++ {
		// Check if client disconnected
		select {
		case <-r.Context().Done():
			fmt.Println("NDJSON client disconnected")
			return
		default:
		}

		// Create and send JSON object
		data := map[string]interface{}{
			"id":        i,
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   fmt.Sprintf("Data record %d", i),
			"value":     i * 100,
		}

		if err := encoder.Encode(data); err != nil {
			fmt.Printf("Error encoding JSON: %v\n", err)
			return
		}

		// Flush after each JSON object
		flusher.Flush()

		time.Sleep(750 * time.Millisecond)
	}

	fmt.Println("NDJSON streaming completed")
}
