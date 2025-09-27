package httpmanager

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
)

// QueryParams represents a map of query parameters from the URL
type QueryParams map[string][]string

// contextKey is a type for context keys
type contextKey string

// queryParamsKey is the context key for query parameters
const queryParamsKey contextKey = "queryParams"

// RequestKey is the context key for the HTTP request
const RequestKey contextKey = "httpRequest"

// GetQueryParams extracts query parameters from the context
func GetQueryParams(ctx context.Context) QueryParams {
	if params, ok := ctx.Value(queryParamsKey).(QueryParams); ok {
		return params
	}
	return QueryParams{}
}

// Get returns the first value for the given query parameter key
func (q QueryParams) Get(key string) string {
	if values, ok := q[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetAll returns all values for the given query parameter key
func (q QueryParams) GetAll(key string) []string {
	if values, ok := q[key]; ok {
		return values
	}
	return []string{}
}

// GetHeader returns a single header value from the context for the given key
func GetHeader(ctx context.Context, key string) string {
	if req, ok := ctx.Value(RequestKey).(*http.Request); ok {
		return req.Header.Get(key)
	}
	return ""
}

// GetHeaders returns all headers from the context
func GetHeaders(ctx context.Context) http.Header {
	if req, ok := ctx.Value(RequestKey).(*http.Request); ok {
		return req.Header
	}
	return http.Header{}
}

// BindQueryParams automatically binds query parameters to a struct using reflection.
// The struct fields should have a "query" tag to specify the query parameter name.
// Supported types: string, int, int64, bool, []string, []int, []int64, []bool
//
// Example usage:
//   type QueryRequest struct {
//       Name     string   `query:"name"`
//       Age      int      `query:"age"`
//       Active   bool     `query:"active"`
//       Tags     []string `query:"tags"`
//   }
//
//   var req QueryRequest
//   err := httpmanager.BindQueryParams(ctx, &req)
func BindQueryParams(ctx context.Context, dst interface{}) error {
	if dst == nil {
		return nil
	}

	queryParams := GetQueryParams(ctx)
	if len(queryParams) == 0 {
		return nil
	}

	// Get the value and type of the destination
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return nil // Must be a pointer to struct
	}

	dstValue = dstValue.Elem()
	if dstValue.Kind() != reflect.Struct {
		return nil // Must point to a struct
	}

	dstType := dstValue.Type()

	// Iterate through struct fields
	for i := 0; i < dstType.NumField(); i++ {
		field := dstType.Field(i)
		fieldValue := dstValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the query tag
		queryTag := field.Tag.Get("query")
		if queryTag == "" {
			continue
		}

		// Get query parameter values
		paramValues := queryParams.GetAll(queryTag)
		if len(paramValues) == 0 {
			continue
		}

		// Bind based on field type
		if err := bindFieldValue(fieldValue, paramValues); err != nil {
			continue // Skip fields that can't be bound
		}
	}

	return nil
}

// bindFieldValue binds query parameter values to a struct field based on its type
func bindFieldValue(fieldValue reflect.Value, paramValues []string) error {
	fieldType := fieldValue.Type()

	switch fieldType.Kind() {
	case reflect.String:
		if len(paramValues) > 0 {
			fieldValue.SetString(paramValues[0])
		}

	case reflect.Int, reflect.Int64:
		if len(paramValues) > 0 {
			if val, err := strconv.ParseInt(paramValues[0], 10, 64); err == nil {
				fieldValue.SetInt(val)
			}
		}

	case reflect.Bool:
		if len(paramValues) > 0 {
			if val, err := strconv.ParseBool(paramValues[0]); err == nil {
				fieldValue.SetBool(val)
			}
		}

	case reflect.Slice:
		switch fieldType.Elem().Kind() {
		case reflect.String:
			fieldValue.Set(reflect.ValueOf(paramValues))

		case reflect.Int, reflect.Int64:
			intSlice := make([]int64, 0, len(paramValues))
			for _, paramValue := range paramValues {
				if val, err := strconv.ParseInt(paramValue, 10, 64); err == nil {
					intSlice = append(intSlice, val)
				}
			}
			if fieldType.Elem().Kind() == reflect.Int {
				intSliceInt := make([]int, len(intSlice))
				for i, v := range intSlice {
					intSliceInt[i] = int(v)
				}
				fieldValue.Set(reflect.ValueOf(intSliceInt))
			} else {
				fieldValue.Set(reflect.ValueOf(intSlice))
			}

		case reflect.Bool:
			boolSlice := make([]bool, 0, len(paramValues))
			for _, paramValue := range paramValues {
				if val, err := strconv.ParseBool(paramValue); err == nil {
					boolSlice = append(boolSlice, val)
				}
			}
			fieldValue.Set(reflect.ValueOf(boolSlice))
		}
	}

	return nil
}
