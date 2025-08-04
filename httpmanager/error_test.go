package httpmanager

import (
	"errors"
	"testing"
)

func TestCustomError_Error(t *testing.T) {
	// Create a new CustomError
	originalErr := errors.New("original error")
	customErr := &CustomError{
		Err:        originalErr,
		Code:       "E123",
		Title:      "Test Error",
		Desc:       "This is a test error",
		StatusCode: 400,
	}

	// Test that Error() returns the original error's message
	if customErr.Error() != originalErr.Error() {
		t.Errorf("Expected error message %q, got %q", originalErr.Error(), customErr.Error())
	}
}

func TestIsCustomError_WithCustomError(t *testing.T) {
	// Create a new CustomError
	originalErr := errors.New("original error")
	customErr := &CustomError{
		Err:        originalErr,
		Code:       "E123",
		Title:      "Test Error",
		Desc:       "This is a test error",
		StatusCode: 400,
	}

	// Test IsCustomError with a CustomError
	resultErr, ok := IsCustomError(customErr)
	if !ok {
		t.Error("Expected IsCustomError to return true for a CustomError")
	}

	// Verify that the returned error is the same as the original
	if resultErr != customErr {
		t.Errorf("Expected IsCustomError to return the original error, got a different error")
	}
}

func TestIsCustomError_WithRegularError(t *testing.T) {
	// Create a regular error
	regularErr := errors.New("regular error")

	// Test IsCustomError with a regular error
	resultErr, ok := IsCustomError(regularErr)
	if ok {
		t.Error("Expected IsCustomError to return false for a regular error")
	}

	// Verify that the returned error is nil
	if resultErr != nil {
		t.Errorf("Expected IsCustomError to return nil for a regular error, got %v", resultErr)
	}
}

func TestCustomError_Fields(t *testing.T) {
	// Create a new CustomError with specific field values
	originalErr := errors.New("original error")
	code := "E123"
	title := "Test Error"
	desc := "This is a test error"
	statusCode := 400

	customErr := &CustomError{
		Err:        originalErr,
		Code:       code,
		Title:      title,
		Desc:       desc,
		StatusCode: statusCode,
	}

	// Verify that all fields have the expected values
	if customErr.Err != originalErr {
		t.Errorf("Expected Err to be %v, got %v", originalErr, customErr.Err)
	}
	if customErr.Code != code {
		t.Errorf("Expected Code to be %q, got %q", code, customErr.Code)
	}
	if customErr.Title != title {
		t.Errorf("Expected Title to be %q, got %q", title, customErr.Title)
	}
	if customErr.Desc != desc {
		t.Errorf("Expected Desc to be %q, got %q", desc, customErr.Desc)
	}
	if customErr.StatusCode != statusCode {
		t.Errorf("Expected StatusCode to be %d, got %d", statusCode, customErr.StatusCode)
	}
}
