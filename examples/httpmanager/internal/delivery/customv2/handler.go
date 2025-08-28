package customv2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

type Handler struct{}

// Custom error types for different scenarios
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

type BusinessError struct {
	Code    string
	Message string
	Reason  string
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("business error [%s]: %s", e.Code, e.Message)
}

type SystemError struct {
	Service string
	Message string
}

func (e SystemError) Error() string {
	return fmt.Sprintf("system error in service '%s': %s", e.Service, e.Message)
}

// NewHandler demonstrates CustomErrorV2 usage with different error types and custom response structures
func NewHandler() *httpmanager.Handler[ProcessOrderRequest, ProcessOrderResponse] {
	return httpmanager.NewHandler(
		http.MethodPost,
		func(ctx context.Context, req *ProcessOrderRequest) (*ProcessOrderResponse, error) {
			requestID := generateRequestID()
			
			// 1. Validation errors with CustomErrorV2 - 400 status
			if err := validateOrder(req); err != nil {
				var validationErr ValidationError
				if errors.As(err, &validationErr) {
					return nil, &httpmanager.CustomErrorV2[ValidationErrorResponse]{
						Err:        err, // Preserve original error
						StatusCode: http.StatusBadRequest,
						Body: ValidationErrorResponse{
							Type:      "validation_error",
							Message:   validationErr.Message,
							Field:     validationErr.Field,
							Value:     validationErr.Value,
							Code:      "FIELD_VALIDATION_FAILED",
							RequestID: requestID,
							Details: map[string]string{
								"hint": "Please check the field value and try again",
								"docs": "https://api.example.com/docs/validation",
							},
						},
					}
				}
			}

			// 2. Business logic errors with CustomErrorV2 - 422 status
			if err := processBusinessLogic(req); err != nil {
				var businessErr BusinessError
				if errors.As(err, &businessErr) {
					return nil, &httpmanager.CustomErrorV2[BusinessErrorResponse]{
						Err:        err, // Preserve original error
						StatusCode: http.StatusUnprocessableEntity,
						Body: BusinessErrorResponse{
							Type:      "business_error",
							Code:      businessErr.Code,
							Message:   businessErr.Message,
							Reason:    businessErr.Reason,
							Timestamp: time.Now().Format(time.RFC3339),
							RequestID: requestID,
							Suggestions: []string{
								"Verify customer account status",
								"Check payment method validity",
								"Ensure sufficient balance",
							},
							Metadata: map[string]interface{}{
								"customer_id":    req.CustomerID,
								"payment_type":   req.PaymentType,
								"attempted_amount": req.Amount,
							},
						},
					}
				}
			}

			// 3. System/internal errors with CustomErrorV2 - 500 status
			if err := processSystemOperation(req); err != nil {
				var systemErr SystemError
				if errors.As(err, &systemErr) {
					return nil, &httpmanager.CustomErrorV2[SystemErrorResponse]{
						Err:        err, // Preserve original error
						StatusCode: http.StatusInternalServerError,
						Body: SystemErrorResponse{
							Type:      "system_error",
							Code:      "INTERNAL_SERVICE_ERROR",
							Message:   "An internal error occurred while processing your request",
							RequestID: requestID,
							Timestamp: time.Now().Format(time.RFC3339),
							Service:   systemErr.Service,
							TraceID:   generateTraceID(),
						},
					}
				}
			}

			// Success response
			return &ProcessOrderResponse{
				OrderID:       req.OrderID,
				Status:        "processed",
				TransactionID: generateTransactionID(),
				Message:       fmt.Sprintf("Order %s processed successfully", req.OrderID),
			}, nil
		},
	)
}

// Validation logic
func validateOrder(req *ProcessOrderRequest) error {
	if strings.TrimSpace(req.OrderID) == "" {
		return ValidationError{
			Field:   "order_id",
			Value:   req.OrderID,
			Message: "order_id is required and cannot be empty",
		}
	}

	if strings.TrimSpace(req.CustomerID) == "" {
		return ValidationError{
			Field:   "customer_id",
			Value:   req.CustomerID,
			Message: "customer_id is required and cannot be empty",
		}
	}

	if req.Amount <= 0 {
		return ValidationError{
			Field:   "amount",
			Value:   req.Amount,
			Message: "amount must be greater than 0",
		}
	}

	if req.Amount > 10000 {
		return ValidationError{
			Field:   "amount",
			Value:   req.Amount,
			Message: "amount cannot exceed 10,000",
		}
	}

	validPaymentTypes := []string{"credit_card", "debit_card", "bank_transfer", "digital_wallet"}
	isValid := false
	for _, validType := range validPaymentTypes {
		if req.PaymentType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return ValidationError{
			Field:   "payment_type",
			Value:   req.PaymentType,
			Message: "payment_type must be one of: " + strings.Join(validPaymentTypes, ", "),
		}
	}

	return nil
}

// Business logic processing
func processBusinessLogic(req *ProcessOrderRequest) error {
	// Simulate business logic errors based on order characteristics
	switch {
	case req.CustomerID == "blocked_customer":
		return BusinessError{
			Code:    "CUSTOMER_BLOCKED",
			Message: "Customer account is blocked",
			Reason:  "Account has been temporarily suspended due to suspicious activity",
		}
	case req.PaymentType == "credit_card" && req.Amount > 5000:
		return BusinessError{
			Code:    "CREDIT_LIMIT_EXCEEDED",
			Message: "Transaction amount exceeds credit limit",
			Reason:  "Available credit limit is insufficient for this transaction",
		}
	case strings.HasPrefix(req.OrderID, "failed_"):
		return BusinessError{
			Code:    "PAYMENT_DECLINED",
			Message: "Payment was declined by the payment processor",
			Reason:  "Insufficient funds or invalid payment method",
		}
	}

	return nil
}

// System operation processing
func processSystemOperation(req *ProcessOrderRequest) error {
	// Simulate system errors based on order characteristics
	switch {
	case strings.Contains(req.OrderID, "db_error"):
		return SystemError{
			Service: "database",
			Message: "database connection timeout",
		}
	case strings.Contains(req.OrderID, "payment_error"):
		return SystemError{
			Service: "payment_gateway",
			Message: "payment service unavailable",
		}
	case strings.Contains(req.OrderID, "notification_error"):
		return SystemError{
			Service: "notification_service",
			Message: "failed to send confirmation email",
		}
	}

	return nil
}

// Utility functions
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func generateTransactionID() string {
	return fmt.Sprintf("txn_%d", time.Now().UnixNano())
}