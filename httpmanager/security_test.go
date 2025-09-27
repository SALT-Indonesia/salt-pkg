package httpmanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBindQueryParams_SecurityAnalysis tests the security aspects of BindQueryParams
func TestBindQueryParams_SecurityAnalysis(t *testing.T) {
	t.Run("reflection safety - only exported fields are settable", func(t *testing.T) {
		type TestStruct struct {
			PublicField  string `query:"public"`
			privateField string `query:"private"` // Unexported field
		}

		params := QueryParams{
			"public":  []string{"public_value"},
			"private": []string{"private_value"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		assert.Equal(t, "public_value", test.PublicField)
		assert.Equal(t, "", test.privateField) // Should remain empty due to CanSet() check
	})

	t.Run("malicious input - extremely long strings", func(t *testing.T) {
		type TestStruct struct {
			Name string `query:"name"`
		}

		// Create a very long string (1MB)
		longString := make([]byte, 1024*1024)
		for i := range longString {
			longString[i] = 'A'
		}

		params := QueryParams{
			"name": []string{string(longString)},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		assert.Equal(t, string(longString), test.Name) // Should handle without error
	})

	t.Run("malicious input - special characters and unicode", func(t *testing.T) {
		type TestStruct struct {
			Name string `query:"name"`
		}

		maliciousInputs := []string{
			"<script>alert('xss')</script>",
			"'; DROP TABLE users; --",
			"../../etc/passwd",
			"\x00\x01\x02",
			"🚀💻🔥", // Unicode emojis
			"null\x00byte",
		}

		for _, maliciousInput := range maliciousInputs {
			params := QueryParams{
				"name": []string{maliciousInput},
			}
			ctx := context.WithValue(context.Background(), queryParamsKey, params)

			var test TestStruct
			err := BindQueryParams(ctx, &test)

			assert.NoError(t, err)
			assert.Equal(t, maliciousInput, test.Name) // Should preserve input as-is without injection
		}
	})

	t.Run("integer overflow protection", func(t *testing.T) {
		type TestStruct struct {
			Value int `query:"value"`
		}

		params := QueryParams{
			"value": []string{"9223372036854775808"}, // Larger than max int64
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		assert.Equal(t, 0, test.Value) // Should remain zero due to parsing error
	})

	t.Run("malicious slice inputs", func(t *testing.T) {
		type TestStruct struct {
			Values []string `query:"values"`
		}

		// Create many values to test potential DoS
		manyValues := make([]string, 10000)
		for i := range manyValues {
			manyValues[i] = "value"
		}

		params := QueryParams{
			"values": manyValues,
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		assert.Equal(t, manyValues, test.Values) // Should handle large slices
	})

	t.Run("nil pointer protection", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"test"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		// Test with nil pointer
		err := BindQueryParams(ctx, nil)
		assert.NoError(t, err) // Should handle gracefully

		// Test with pointer to nil
		var test *struct {
			Name string `query:"name"`
		}
		err = BindQueryParams(ctx, test)
		assert.NoError(t, err) // Should handle gracefully
	})

	t.Run("type confusion attacks", func(t *testing.T) {
		type TestStruct struct {
			Value interface{} `query:"value"` // Interface{} field
		}

		params := QueryParams{
			"value": []string{"malicious"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		// interface{} fields should be skipped since they're not supported types
		assert.Nil(t, test.Value)
	})

	t.Run("reflection limits - non-struct destination", func(t *testing.T) {
		params := QueryParams{
			"value": []string{"test"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		// Test with non-struct types
		var stringVar string
		err := BindQueryParams(ctx, &stringVar)
		assert.NoError(t, err) // Should handle gracefully

		var intVar int
		err = BindQueryParams(ctx, &intVar)
		assert.NoError(t, err) // Should handle gracefully

		var sliceVar []string
		err = BindQueryParams(ctx, &sliceVar)
		assert.NoError(t, err) // Should handle gracefully
	})

	t.Run("memory exhaustion protection", func(t *testing.T) {
		type TestStruct struct {
			Values []int `query:"values"`
		}

		// Test with many invalid integer values
		manyInvalidValues := make([]string, 10000)
		for i := range manyInvalidValues {
			manyInvalidValues[i] = "invalid_int"
		}

		params := QueryParams{
			"values": manyInvalidValues,
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		assert.NoError(t, err)
		assert.Empty(t, test.Values) // Should result in empty slice due to parsing failures
	})

	t.Run("no information disclosure in errors", func(t *testing.T) {
		type TestStruct struct {
			Value int `query:"value"`
		}

		params := QueryParams{
			"value": []string{"invalid"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var test TestStruct
		err := BindQueryParams(ctx, &test)

		// Should not return error even with invalid input - graceful degradation
		assert.NoError(t, err)
		assert.Equal(t, 0, test.Value) // Should remain zero value
	})
}