package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the OMS system
type User struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	Orders    []Order        `gorm:"foreignKey:UserID"` // Ensure the foreign key is correctly set

}

type UserOrder struct {
	UserID  int `json:"user_id"`
	OrderID int `json:"order_id"`
}



type UserResponse struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}



type UserOrderResponse struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	OrderResponse []OrderResponse `json:"order_response"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at"`
}

type OrderResponse struct {
	ID         int            `json:"id"`
	TotalPrice float64        `json:"total_price"`
	Status     string         `json:"status"`
	FinalPrice float64        `json:"final_price"` // Total price after applying discounts
	Items      []ItemResponse `json:"items"`       // List of items in the order
}

// OrderItem represents an item in an order
type ItemResponse struct {
	ItemID   int     `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}
