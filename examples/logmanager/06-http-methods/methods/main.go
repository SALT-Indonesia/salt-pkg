package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "trace_id"

type Resource struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

var resources = make(map[string]*Resource)

func main() {
	// Initialize with sample data
	initSampleData()

	app := logmanager.NewApplication(
		logmanager.WithAppName("http-methods"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders("X-Custom-Header", "Authorization"),
		logmanager.WithDebug(),
	)

	r := gin.Default()
	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	// Standard REST endpoints demonstrating all HTTP methods
	api := r.Group("/api/v1")
	{
		// GET - Retrieve resources
		api.GET("/resources", getAllResources)
		api.GET("/resources/:id", getResourceByID)

		// POST - Create new resource
		api.POST("/resources", createResource)

		// PUT - Update entire resource
		api.PUT("/resources/:id", updateResource)

		// PATCH - Partial update
		api.PATCH("/resources/:id", patchResource)

		// DELETE - Remove resource
		api.DELETE("/resources/:id", deleteResource)

		// HEAD - Get headers only
		api.HEAD("/resources/:id", headResource)

		// OPTIONS - Get allowed methods
		api.OPTIONS("/resources", optionsResource)
		api.OPTIONS("/resources/:id", optionsResource)
	}

	// Additional HTTP method demonstrations
	customMethods := r.Group("/custom")
	{
		// Custom verbs (not REST standard but valid HTTP)
		customMethods.Handle("TRACE", "/trace", traceRequest)
		customMethods.Handle("CONNECT", "/connect", connectRequest)
	}

	fmt.Println("HTTP Methods server running at http://localhost:8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /api/v1/resources")
	fmt.Println("  GET    /api/v1/resources/:id")
	fmt.Println("  POST   /api/v1/resources")
	fmt.Println("  PUT    /api/v1/resources/:id")
	fmt.Println("  PATCH  /api/v1/resources/:id")
	fmt.Println("  DELETE /api/v1/resources/:id")
	fmt.Println("  HEAD   /api/v1/resources/:id")
	fmt.Println("  OPTIONS /api/v1/resources")

	log.Fatal(r.Run(":8080"))
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

func initSampleData() {
	resources["1"] = &Resource{
		ID:          "1",
		Name:        "Sample Resource 1",
		Description: "This is a sample resource for testing",
		Metadata:    map[string]interface{}{"category": "test", "priority": "high"},
	}
	resources["2"] = &Resource{
		ID:          "2",
		Name:        "Sample Resource 2",
		Description: "Another sample resource",
		Metadata:    map[string]interface{}{"category": "demo", "priority": "medium"},
	}
}

// GET /api/v1/resources - Get all resources
func getAllResources(c *gin.Context) {
	logSegment(c, "get-all-resources", func() {
		result := make([]*Resource, 0, len(resources))
		for _, resource := range resources {
			result = append(result, resource)
		}

		c.Header("X-Total-Count", fmt.Sprintf("%d", len(result)))
		c.JSON(http.StatusOK, gin.H{
			"data":  result,
			"count": len(result),
		})
	})
}

// GET /api/v1/resources/:id - Get specific resource
func getResourceByID(c *gin.Context) {
	id := c.Param("id")

	logSegment(c, "get-resource-by-id", func() {
		resource, exists := resources[id]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resource not found",
				"id":    id,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": resource,
		})
	})
}

// POST /api/v1/resources - Create new resource
func createResource(c *gin.Context) {
	var newResource Resource

	if err := c.ShouldBindJSON(&newResource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	logSegment(c, "create-resource", func() {
		newResource.ID = uuid.NewString()
		resources[newResource.ID] = &newResource

		c.Header("Location", fmt.Sprintf("/api/v1/resources/%s", newResource.ID))
		c.JSON(http.StatusCreated, gin.H{
			"data": newResource,
			"message": "Resource created successfully",
		})
	})
}

// PUT /api/v1/resources/:id - Update entire resource
func updateResource(c *gin.Context) {
	id := c.Param("id")
	var updatedResource Resource

	if err := c.ShouldBindJSON(&updatedResource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	logSegment(c, "update-resource", func() {
		if _, exists := resources[id]; !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resource not found",
				"id":    id,
			})
			return
		}

		updatedResource.ID = id
		resources[id] = &updatedResource

		c.JSON(http.StatusOK, gin.H{
			"data": updatedResource,
			"message": "Resource updated successfully",
		})
	})
}

// PATCH /api/v1/resources/:id - Partial update
func patchResource(c *gin.Context) {
	id := c.Param("id")
	var patch map[string]interface{}

	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	logSegment(c, "patch-resource", func() {
		resource, exists := resources[id]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resource not found",
				"id":    id,
			})
			return
		}

		// Apply partial updates
		if name, ok := patch["name"].(string); ok {
			resource.Name = name
		}
		if desc, ok := patch["description"].(string); ok {
			resource.Description = desc
		}
		if metadata, ok := patch["metadata"].(map[string]interface{}); ok {
			if resource.Metadata == nil {
				resource.Metadata = make(map[string]interface{})
			}
			for k, v := range metadata {
				resource.Metadata[k] = v
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": resource,
			"message": "Resource patched successfully",
		})
	})
}

// DELETE /api/v1/resources/:id - Delete resource
func deleteResource(c *gin.Context) {
	id := c.Param("id")

	logSegment(c, "delete-resource", func() {
		if _, exists := resources[id]; !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resource not found",
				"id":    id,
			})
			return
		}

		delete(resources, id)

		c.JSON(http.StatusOK, gin.H{
			"message": "Resource deleted successfully",
			"id":      id,
		})
	})
}

// HEAD /api/v1/resources/:id - Get headers only
func headResource(c *gin.Context) {
	id := c.Param("id")

	logSegment(c, "head-resource", func() {
		resource, exists := resources[id]
		if !exists {
			c.Status(http.StatusNotFound)
			return
		}

		// Set headers with resource metadata
		c.Header("X-Resource-ID", resource.ID)
		c.Header("X-Resource-Name", resource.Name)
		c.Header("Content-Type", "application/json")
		c.Header("Last-Modified", "Wed, 21 Oct 2024 07:28:00 GMT")

		c.Status(http.StatusOK)
	})
}

// OPTIONS /api/v1/resources - Get allowed methods
func optionsResource(c *gin.Context) {
	logSegment(c, "options-resource", func() {
		c.Header("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Accept", "application/json")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Trace-ID")

		c.Status(http.StatusOK)
	})
}

// TRACE /custom/trace - Echo request for debugging
func traceRequest(c *gin.Context) {
	logSegment(c, "trace-request", func() {
		c.JSON(http.StatusOK, gin.H{
			"method":  c.Request.Method,
			"url":     c.Request.URL.String(),
			"headers": c.Request.Header,
			"trace":   "Request traced successfully",
		})
	})
}

// CONNECT /custom/connect - Tunnel establishment (demo)
func connectRequest(c *gin.Context) {
	logSegment(c, "connect-request", func() {
		c.JSON(http.StatusOK, gin.H{
			"method":  c.Request.Method,
			"message": "CONNECT method demonstration",
			"note":    "In real scenarios, CONNECT is used for tunneling",
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