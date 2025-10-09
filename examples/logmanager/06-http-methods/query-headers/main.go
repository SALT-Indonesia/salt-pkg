package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "trace_id"

type QueryParams struct {
	Page     int      `form:"page" json:"page"`
	Limit    int      `form:"limit" json:"limit"`
	Sort     string   `form:"sort" json:"sort"`
	Filter   string   `form:"filter" json:"filter"`
	Fields   []string `form:"fields" json:"fields"`
	Include  []string `form:"include" json:"include"`
	Search   string   `form:"q" json:"search"`
	Category string   `form:"category" json:"category"`
}

type HeaderInfo struct {
	Authorization   string `json:"authorization,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	ContentType    string `json:"content_type,omitempty"`
	Accept         string `json:"accept,omitempty"`
	AcceptLanguage string `json:"accept_language,omitempty"`
	XForwardedFor  string `json:"x_forwarded_for,omitempty"`
	XRealIP        string `json:"x_real_ip,omitempty"`
	Custom         map[string]string `json:"custom_headers,omitempty"`
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-query-headers"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders(
			"Authorization", "X-API-Key", "X-Custom-Header",
			"X-Client-Version", "X-Request-ID", "User-Agent",
		),
		logmanager.WithDebug(),
	)

	r := gin.Default()
	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	// Query parameter examples
	query := r.Group("/query")
	{
		// Simple query parameters
		query.GET("/simple", handleSimpleQuery)

		// Complex query parameters with arrays
		query.GET("/complex", handleComplexQuery)

		// Pagination with query params
		query.GET("/paginated", handlePaginatedQuery)

		// Search with filters
		query.GET("/search", handleSearchQuery)

		// Nested/encoded query parameters
		query.GET("/nested", handleNestedQuery)
	}

	// Header examples
	headers := r.Group("/headers")
	{
		// Standard HTTP headers
		headers.GET("/standard", handleStandardHeaders)

		// Authentication headers
		headers.GET("/auth", handleAuthHeaders)

		// Custom headers
		headers.GET("/custom", handleCustomHeaders)

		// Content negotiation headers
		headers.GET("/negotiation", handleContentNegotiation)

		// Forwarding headers (proxy scenarios)
		headers.GET("/forwarding", handleForwardingHeaders)
	}

	// Combined query + headers
	combined := r.Group("/combined")
	{
		// API with both query params and headers
		combined.GET("/api", handleCombinedAPI)

		// RESTful resource with filtering and auth
		combined.GET("/resources", handleResourcesAPI)
	}

	// Special scenarios
	special := r.Group("/special")
	{
		// Case-insensitive headers
		special.GET("/case-insensitive", handleCaseInsensitiveHeaders)

		// Multiple values for same parameter
		special.GET("/multi-value", handleMultiValueParams)

		// URL encoding scenarios
		special.GET("/encoded", handleEncodedParams)
	}

	fmt.Println("HTTP Query & Headers server running at http://localhost:8082")
	fmt.Println("Available endpoints:")
	fmt.Println("Query Parameters:")
	fmt.Println("  GET /query/simple        - Basic query params")
	fmt.Println("  GET /query/complex       - Complex arrays and filters")
	fmt.Println("  GET /query/paginated     - Pagination examples")
	fmt.Println("  GET /query/search        - Search with filters")
	fmt.Println("Headers:")
	fmt.Println("  GET /headers/standard    - Standard HTTP headers")
	fmt.Println("  GET /headers/auth        - Authentication headers")
	fmt.Println("  GET /headers/custom      - Custom headers")
	fmt.Println("Combined:")
	fmt.Println("  GET /combined/api        - Query params + headers")

	log.Fatal(r.Run(":8082"))
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

// Simple query parameters: ?name=john&age=30&active=true
func handleSimpleQuery(c *gin.Context) {
	logSegment(c, "handle-simple-query", func() {
		name := c.Query("name")
		age := c.Query("age")
		active := c.Query("active")
		city := c.DefaultQuery("city", "unknown")

		c.JSON(http.StatusOK, gin.H{
			"message": "Simple query parameters processed",
			"params": map[string]string{
				"name":   name,
				"age":    age,
				"active": active,
				"city":   city,
			},
		})
	})
}

// Complex query: ?fields[]=name&fields[]=email&include[]=profile&sort=name:asc,age:desc
func handleComplexQuery(c *gin.Context) {
	logSegment(c, "handle-complex-query", func() {
		var params QueryParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid query parameters",
				"details": err.Error(),
			})
			return
		}

		// Parse complex parameters manually for more control
		fields := c.QueryArray("fields")
		include := c.QueryArray("include")
		tags := c.QueryArray("tags")

		c.JSON(http.StatusOK, gin.H{
			"message": "Complex query parameters processed",
			"params":  params,
			"arrays": map[string][]string{
				"fields":  fields,
				"include": include,
				"tags":    tags,
			},
		})
	})
}

// Pagination: ?page=2&limit=20&sort=created_at:desc
func handlePaginatedQuery(c *gin.Context) {
	logSegment(c, "handle-paginated-query", func() {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		sort := c.DefaultQuery("sort", "id:asc")

		// Validate pagination
		if page < 1 {
			page = 1
		}
		if limit > 100 {
			limit = 100
		}

		offset := (page - 1) * limit

		c.JSON(http.StatusOK, gin.H{
			"message": "Pagination parameters processed",
			"pagination": map[string]interface{}{
				"page":   page,
				"limit":  limit,
				"offset": offset,
				"sort":   sort,
			},
			"meta": map[string]interface{}{
				"total":       1000, // Example total
				"total_pages": (1000 + limit - 1) / limit,
				"has_next":    page*limit < 1000,
				"has_prev":    page > 1,
			},
		})
	})
}

// Search: ?q=search+term&category=tech&price_min=10&price_max=100
func handleSearchQuery(c *gin.Context) {
	logSegment(c, "handle-search-query", func() {
		query := c.Query("q")
		category := c.Query("category")
		priceMin := c.Query("price_min")
		priceMax := c.Query("price_max")
		dateFrom := c.Query("date_from")
		dateTo := c.Query("date_to")

		filters := make(map[string]string)
		if category != "" {
			filters["category"] = category
		}
		if priceMin != "" {
			filters["price_min"] = priceMin
		}
		if priceMax != "" {
			filters["price_max"] = priceMax
		}
		if dateFrom != "" {
			filters["date_from"] = dateFrom
		}
		if dateTo != "" {
			filters["date_to"] = dateTo
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Search query processed",
			"search": map[string]interface{}{
				"query":   query,
				"filters": filters,
				"results": "Sample search results...",
			},
		})
	})
}

// Nested parameters: ?user[name]=john&user[profile][age]=30
func handleNestedQuery(c *gin.Context) {
	logSegment(c, "handle-nested-query", func() {
		// Manual parsing of nested parameters
		params := make(map[string]interface{})

		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				if strings.Contains(key, "[") && strings.Contains(key, "]") {
					// This is a nested parameter - simplified parsing
					params[key] = values[0]
				} else {
					params[key] = values[0]
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Nested query parameters processed",
			"params":  params,
		})
	})
}

// Standard HTTP headers
func handleStandardHeaders(c *gin.Context) {
	logSegment(c, "handle-standard-headers", func() {
		headers := HeaderInfo{
			UserAgent:      c.GetHeader("User-Agent"),
			ContentType:    c.GetHeader("Content-Type"),
			Accept:         c.GetHeader("Accept"),
			AcceptLanguage: c.GetHeader("Accept-Language"),
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Standard headers processed",
			"headers": headers,
			"all_headers": c.Request.Header,
		})
	})
}

// Authentication headers
func handleAuthHeaders(c *gin.Context) {
	logSegment(c, "handle-auth-headers", func() {
		auth := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")
		clientID := c.GetHeader("X-Client-ID")
		signature := c.GetHeader("X-Signature")

		// Parse Authorization header
		var authType, token string
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 {
				authType = parts[0]
				token = parts[1]
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Authentication headers processed",
			"auth": map[string]string{
				"type":       authType,
				"token":      maskToken(token),
				"api_key":    maskToken(apiKey),
				"client_id":  clientID,
				"signature":  maskToken(signature),
			},
		})
	})
}

// Custom headers
func handleCustomHeaders(c *gin.Context) {
	logSegment(c, "handle-custom-headers", func() {
		customHeaders := make(map[string]string)

		// Look for custom headers (X- prefix)
		for name, values := range c.Request.Header {
			if strings.HasPrefix(name, "X-") && len(values) > 0 {
				customHeaders[name] = values[0]
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Custom headers processed",
			"custom_headers": customHeaders,
			"timestamp": time.Now().Unix(),
		})
	})
}

// Content negotiation headers
func handleContentNegotiation(c *gin.Context) {
	logSegment(c, "handle-content-negotiation", func() {
		accept := c.GetHeader("Accept")
		acceptLang := c.GetHeader("Accept-Language")
		acceptEncoding := c.GetHeader("Accept-Encoding")

		// Simple content negotiation
		var responseType string
		if strings.Contains(accept, "application/json") {
			responseType = "json"
		} else if strings.Contains(accept, "application/xml") {
			responseType = "xml"
		} else if strings.Contains(accept, "text/plain") {
			responseType = "text"
		} else {
			responseType = "default"
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Content negotiation processed",
			"negotiation": map[string]string{
				"accept":          accept,
				"accept_language": acceptLang,
				"accept_encoding": acceptEncoding,
				"response_type":   responseType,
			},
		})
	})
}

// Forwarding headers (proxy scenarios)
func handleForwardingHeaders(c *gin.Context) {
	logSegment(c, "handle-forwarding-headers", func() {
		headers := HeaderInfo{
			XForwardedFor: c.GetHeader("X-Forwarded-For"),
			XRealIP:       c.GetHeader("X-Real-IP"),
		}

		// Additional forwarding headers
		forwardingHeaders := map[string]string{
			"X-Forwarded-Proto": c.GetHeader("X-Forwarded-Proto"),
			"X-Forwarded-Host":  c.GetHeader("X-Forwarded-Host"),
			"X-Forwarded-Port":  c.GetHeader("X-Forwarded-Port"),
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Forwarding headers processed",
			"headers": headers,
			"forwarding": forwardingHeaders,
			"client_ip": c.ClientIP(),
		})
	})
}

// Combined API with query params and headers
func handleCombinedAPI(c *gin.Context) {
	logSegment(c, "handle-combined-api", func() {
		// Query parameters
		var params QueryParams
		c.ShouldBindQuery(&params)

		// Headers
		auth := c.GetHeader("Authorization")
		userAgent := c.GetHeader("User-Agent")
		apiKey := c.GetHeader("X-API-Key")

		c.JSON(http.StatusOK, gin.H{
			"message": "Combined query and headers processed",
			"query_params": params,
			"headers": map[string]string{
				"authorization": maskToken(auth),
				"user_agent":    userAgent,
				"api_key":       maskToken(apiKey),
			},
			"timestamp": time.Now().Unix(),
		})
	})
}

// RESTful resources API
func handleResourcesAPI(c *gin.Context) {
	logSegment(c, "handle-resources-api", func() {
		// Authorization check
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Query parameters for filtering
		category := c.Query("category")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		c.JSON(http.StatusOK, gin.H{
			"message": "Resources API processed",
			"filters": map[string]interface{}{
				"category": category,
				"status":   status,
				"page":     page,
				"limit":    limit,
			},
			"auth_status": "authenticated",
			"data": []map[string]interface{}{
				{"id": 1, "name": "Resource 1", "category": category},
				{"id": 2, "name": "Resource 2", "category": category},
			},
		})
	})
}

// Case-insensitive headers
func handleCaseInsensitiveHeaders(c *gin.Context) {
	logSegment(c, "handle-case-insensitive-headers", func() {
		// HTTP headers are case-insensitive
		headers := map[string]string{
			"content-type (lowercase)": c.GetHeader("content-type"),
			"Content-Type (title)":     c.GetHeader("Content-Type"),
			"CONTENT-TYPE (uppercase)": c.GetHeader("CONTENT-TYPE"),
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Case-insensitive headers demonstration",
			"headers": headers,
			"note": "HTTP headers are case-insensitive",
		})
	})
}

// Multiple values for same parameter
func handleMultiValueParams(c *gin.Context) {
	logSegment(c, "handle-multi-value-params", func() {
		// ?tags=tech&tags=golang&tags=api
		tags := c.QueryArray("tags")
		categories := c.QueryArray("categories")

		// Query map for all parameters
		queryMap := c.Request.URL.Query()

		c.JSON(http.StatusOK, gin.H{
			"message": "Multi-value parameters processed",
			"arrays": map[string][]string{
				"tags":       tags,
				"categories": categories,
			},
			"all_params": queryMap,
		})
	})
}

// URL encoded parameters
func handleEncodedParams(c *gin.Context) {
	logSegment(c, "handle-encoded-params", func() {
		// Parameters with special characters that need encoding
		search := c.Query("q")           // e.g., "hello world" -> "hello%20world"
		email := c.Query("email")        // e.g., "user@domain.com" -> "user%40domain.com"
		special := c.Query("special")    // e.g., "a+b=c&d" -> "a%2Bb%3Dc%26d"

		c.JSON(http.StatusOK, gin.H{
			"message": "URL encoded parameters processed",
			"decoded_params": map[string]string{
				"search":  search,
				"email":   email,
				"special": special,
			},
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

// Helper function to mask sensitive tokens
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}