package customv2

type ProcessOrderRequest struct {
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	Amount      float64 `json:"amount"`
	PaymentType string  `json:"payment_type"`
}