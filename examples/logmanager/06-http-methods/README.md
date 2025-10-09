# HTTP Methods & Content Types Examples

Comprehensive demonstration of HTTP methods, content types, query parameters, headers, and advanced HTTP features with logmanager integration.

## 📁 Directory Structure

```
06-http-methods/
├── methods/              # All HTTP methods (GET, POST, PUT, DELETE, PATCH, etc.)
├── content-types/        # Content type variations (JSON, form-data, binary, etc.)
├── query-headers/        # Query parameters and HTTP headers
└── advanced/             # Advanced features (file uploads, streaming)
```

## 🚀 Running the Examples

Each directory contains a standalone server demonstrating specific HTTP features:

```bash
# HTTP Methods (REST API)
cd methods && go run main.go              # Port 8080

# Content Types
cd content-types && go run main.go        # Port 8081

# Query Parameters & Headers
cd query-headers && go run main.go        # Port 8082

# Advanced Features
cd advanced && go run main.go             # Port 8083
```

## 📋 Complete HTTP Methods Coverage

### Standard REST Methods

| Method | Endpoint | Description | Example |
|--------|----------|-------------|---------|
| **GET** | `/api/v1/resources` | Retrieve all resources | `curl http://localhost:8080/api/v1/resources` |
| **GET** | `/api/v1/resources/:id` | Retrieve specific resource | `curl http://localhost:8080/api/v1/resources/1` |
| **POST** | `/api/v1/resources` | Create new resource | `curl -X POST -H "Content-Type: application/json" -d '{"name":"test"}' http://localhost:8080/api/v1/resources` |
| **PUT** | `/api/v1/resources/:id` | Update entire resource | `curl -X PUT -H "Content-Type: application/json" -d '{"name":"updated"}' http://localhost:8080/api/v1/resources/1` |
| **PATCH** | `/api/v1/resources/:id` | Partial update | `curl -X PATCH -H "Content-Type: application/json" -d '{"name":"patched"}' http://localhost:8080/api/v1/resources/1` |
| **DELETE** | `/api/v1/resources/:id` | Delete resource | `curl -X DELETE http://localhost:8080/api/v1/resources/1` |
| **HEAD** | `/api/v1/resources/:id` | Get headers only | `curl -I http://localhost:8080/api/v1/resources/1` |
| **OPTIONS** | `/api/v1/resources` | Get allowed methods | `curl -X OPTIONS http://localhost:8080/api/v1/resources` |

### Additional HTTP Methods

| Method | Endpoint | Description |
|--------|----------|-------------|
| **TRACE** | `/custom/trace` | Echo request for debugging |
| **CONNECT** | `/custom/connect` | Tunnel establishment demo |

## 🔧 Content Types Supported

### Request Content Types

| Content-Type | Endpoint | Description | Example |
|--------------|----------|-------------|---------|
| **application/json** | `/api/v1/json` | JSON payload | `curl -X POST -H "Content-Type: application/json" -d '{"name":"John","age":30}' http://localhost:8081/api/v1/json` |
| **multipart/form-data** | `/api/v1/form-data` | Form with file upload | `curl -X POST -F "name=John" -F "file=@test.txt" http://localhost:8081/api/v1/form-data` |
| **application/x-www-form-urlencoded** | `/api/v1/urlencoded` | URL encoded form | `curl -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "name=John&age=30" http://localhost:8081/api/v1/urlencoded` |
| **text/plain** | `/api/v1/text` | Plain text | `curl -X POST -H "Content-Type: text/plain" -d "Hello World" http://localhost:8081/api/v1/text` |
| **application/octet-stream** | `/api/v1/binary` | Binary data | `curl -X POST -H "Content-Type: application/octet-stream" --data-binary @file.bin http://localhost:8081/api/v1/binary` |
| **application/xml** | `/api/v1/xml` | XML data | `curl -X POST -H "Content-Type: application/xml" -d "<user><name>John</name></user>" http://localhost:8081/api/v1/xml` |

### File Upload Variations

| Type | Endpoint | Description |
|------|----------|-------------|
| **Single File** | `/api/v1/upload` | Single file upload |
| **Multiple Files** | `/api/v1/upload-multiple` | Multiple files at once |
| **Chunked Upload** | `/api/v1/upload-chunked` | Large file in chunks |
| **Base64 Upload** | `/api/v1/upload-base64` | Base64 encoded file |

## 📊 Query Parameters & Headers

### Query Parameter Types

| Type | Example | Description |
|------|---------|-------------|
| **Simple** | `?name=john&age=30` | Basic key-value pairs |
| **Arrays** | `?tags[]=go&tags[]=api` | Multiple values |
| **Pagination** | `?page=2&limit=20&sort=name:asc` | Pagination parameters |
| **Search/Filter** | `?q=search&category=tech&price_min=10` | Search with filters |
| **Nested** | `?user[name]=john&user[age]=30` | Nested parameters |

### Header Categories

| Category | Examples | Usage |
|----------|----------|--------|
| **Standard** | `Content-Type`, `Accept`, `User-Agent` | HTTP protocol headers |
| **Authentication** | `Authorization`, `X-API-Key` | Authentication tokens |
| **Custom** | `X-Request-ID`, `X-Client-Version` | Application-specific |
| **Forwarding** | `X-Forwarded-For`, `X-Real-IP` | Proxy headers |
| **Content Negotiation** | `Accept-Language`, `Accept-Encoding` | Content preferences |

## 🚀 Advanced HTTP Features


### File Upload Features

| Feature | Endpoint | Description |
|---------|----------|-------------|
| **Single Upload** | `/upload/single` | Basic file upload |
| **Multiple Upload** | `/upload/multiple` | Multiple files |
| **Chunked Upload** | `/upload/chunked` | Large files in chunks |
| **Progress Tracking** | `/upload/large` | Upload with progress |
| **Base64 Upload** | `/upload/base64` | Encoded file data |

### Streaming Features

| Feature | Endpoint | Description |
|---------|----------|-------------|
| **Server-Sent Events** | `/stream/events` | Real-time events |
| **Chunked Response** | `/stream/chunked` | Chunked transfer encoding |
| **Data Stream** | `/stream/data` | NDJSON data stream |
| **File Download** | `/stream/download/:file` | Streaming downloads |

### Real-time Features

| Feature | Endpoint | Description |
|---------|----------|-------------|
| **Long Polling** | `/realtime/poll` | Long-lived connections |
| **Webhooks** | `/realtime/webhook` | Webhook receiver |

### Advanced HTTP

| Feature | Endpoint | Description |
|---------|----------|-------------|
| **Conditional Requests** | `/advanced/conditional` | ETags, If-Modified-Since |
| **CORS** | `/advanced/cors` | Cross-origin requests |
| **Compression** | `/advanced/compressed` | Gzip compressed responses |
| **Partial Content** | `/advanced/partial` | Range requests |

## 🔍 Testing Examples

### REST API Testing
```bash
# Create resource
curl -X POST http://localhost:8080/api/v1/resources \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Resource","description":"A test resource"}'

# Get all resources
curl http://localhost:8080/api/v1/resources

# Update resource
curl -X PUT http://localhost:8080/api/v1/resources/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Resource","description":"Updated description"}'

# Partial update
curl -X PATCH http://localhost:8080/api/v1/resources/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Patched Name"}'

# Delete resource
curl -X DELETE http://localhost:8080/api/v1/resources/1
```

### Content Type Testing
```bash
# JSON
curl -X POST http://localhost:8081/api/v1/json \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","age":30}'

# Form data with file
curl -X POST http://localhost:8081/api/v1/form-data \
  -F "name=John" \
  -F "email=john@example.com" \
  -F "avatar=@profile.jpg"

# URL encoded
curl -X POST http://localhost:8081/api/v1/urlencoded \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=John&email=john@example.com&age=30"

# Binary data
curl -X POST http://localhost:8081/api/v1/binary \
  -H "Content-Type: application/octet-stream" \
  --data-binary @data.bin
```

### Query Parameters Testing
```bash
# Simple query
curl "http://localhost:8082/query/simple?name=john&age=30&active=true"

# Complex arrays
curl "http://localhost:8082/query/complex?fields[]=name&fields[]=email&include[]=profile&sort=name:asc"

# Pagination
curl "http://localhost:8082/query/paginated?page=2&limit=20&sort=created_at:desc"

# Search with filters
curl "http://localhost:8082/query/search?q=golang&category=tech&price_min=10&price_max=100"
```

### Headers Testing
```bash
# Authentication headers
curl http://localhost:8082/headers/auth \
  -H "Authorization: Bearer token123" \
  -H "X-API-Key: api_key_123" \
  -H "X-Client-ID: client_123"

# Custom headers
curl http://localhost:8082/headers/custom \
  -H "X-Request-ID: req_123" \
  -H "X-Client-Version: 1.0.0" \
  -H "X-Custom-Header: custom_value"

# Content negotiation
curl http://localhost:8082/headers/negotiation \
  -H "Accept: application/json" \
  -H "Accept-Language: en-US,en;q=0.9" \
  -H "Accept-Encoding: gzip, deflate"
```

### Advanced Features Testing
```bash

# File upload
curl -X POST http://localhost:8083/upload/single \
  -F "file=@document.pdf"

# Server-Sent Events
curl http://localhost:8083/stream/events

# Chunked response
curl http://localhost:8083/stream/chunked

# Conditional request
curl http://localhost:8083/advanced/conditional \
  -H "If-None-Match: \"123456789\""

# Range request
curl http://localhost:8083/advanced/partial \
  -H "Range: bytes=0-499"
```

## 🎯 Key Features Demonstrated

### Logmanager Integration
- **Transaction Tracking**: Each request gets unique transaction
- **Segment Logging**: Database queries, external calls tracked
- **Error Handling**: Automatic error capture and logging
- **Trace ID**: Consistent trace ID across all operations
- **Performance Monitoring**: Request timing and metrics

### HTTP Protocol Features
- **All HTTP Methods**: Complete REST and HTTP method coverage
- **Content Types**: Every major content type handled
- **Headers**: Standard and custom header processing
- **Query Parameters**: Simple to complex parameter handling
- **Status Codes**: Appropriate HTTP status code usage

### Advanced Patterns
- **File Handling**: Various upload/download patterns
- **Streaming**: Real-time data streaming patterns
- **Content Negotiation**: Accept header handling
- **CORS**: Cross-origin request support

## 📚 Learning Path

1. **Start with Methods** (`methods/`) - Learn REST API patterns
2. **Explore Content Types** (`content-types/`) - Understand different data formats
3. **Master Query/Headers** (`query-headers/`) - Handle parameters and headers
4. **Advanced Features** (`advanced/`) - Streaming, file uploads

## 🔗 Related Examples

- [HTTP Servers](../02-http-servers/) - Framework-specific implementations
- [Data Masking](../05-masking/) - Protect sensitive data in requests
- [Basic Usage](../01-basic/) - Core logmanager concepts