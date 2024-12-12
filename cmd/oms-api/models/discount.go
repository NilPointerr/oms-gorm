package models

type Discounts struct {
	SeasonalDiscount    float64 `json:"seasonal_discount"`
	VolumeBasedDiscount float64 `json:"volume_based_discount"`
	LoyaltyDiscount     float64 `json:"loyalty_discount"`
	TotalDiscountAmount float64 `json:"total_discount_amount"`
}

type DiscountRequest struct {
	UserID  int `json:"user_id"`
	OrderID int `json:"order_id"`
}
