package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	"gorm.io/gorm"
)

// var db *sql.DB // Global variable to hold the database connection

// CreateOrder creates a new order and stores it in the database using GORM
func CreateOrder(c *gin.Context, db *gorm.DB) {
	// Check for nil database connection
	if db == nil {
		log.Println("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is not available"})
		return
	}

	// Bind the incoming JSON to the Order model
	var newOrder models.Order
	if err := c.ShouldBindJSON(&newOrder); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Initialize total price and order items
	var totalPrice float64
	var orderItems []models.OrderItem

	// Loop through the OrderItems to calculate the total price
	for _, item := range newOrder.Items {
		var itemRecord models.Item
		// Use GORM's First method to get the item by ID
		if err := db.First(&itemRecord, item.ItemID).Error; err != nil {
			log.Println("Error fetching item for item ID", item.ItemID, ":", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item"})
			return
		}

		// Extract the price from the fetched item record
		price := itemRecord.Price

		// Calculate the total price
		totalPrice += price * float64(item.Quantity)

		// Populate item details, including price
		item.Price = price
		item.OrderID = newOrder.ID // Set OrderID here to link it with the order
		// Add item to the orderItems array
		orderItems = append(orderItems, item)
	}

	// Calculate discounts based on predefined conditions
	discounts := calculateDiscounts(db, newOrder, orderItems)

	// Calculate the final price after applying discounts
	finalPrice := calculateTotalPrice(db, orderItems, discounts)

	// Set the total and final price in the order object
	newOrder.TotalPrice = totalPrice
	newOrder.FinalPrice = finalPrice

	// Insert the new order into the database
	if err := db.Create(&newOrder).Error; err != nil {
		log.Println("Error inserting order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert order"})
		return
	}

	// Insert items into the order_items table using GORM
	for _, item := range orderItems {
		item.OrderID = newOrder.ID
		if err := db.Create(&item).Error; err != nil {
			log.Println("Error inserting order item:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert order item"})
			return
		}
	}

	// Insert user_id and order_id into the userOrder table
	if err := db.Model(&models.UserOrder{}).Create(&models.UserOrder{
		UserID:  newOrder.UserID,
		OrderID: newOrder.ID,
	}).Error; err != nil {
		log.Println("Error inserting into userOrder table:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into userOrder table"})
		return
	}

	// Respond with the created order and its items
	c.JSON(http.StatusOK, gin.H{
			"order": newOrder,
		})
}

func GetOrders(c *gin.Context, db *gorm.DB) {
	var orders []models.Order
	// Fetch orders with GORM, excluding soft-deleted orders
	if err := db.Where("deleted_at IS NULL").Find(&orders).Error; err != nil {
		log.Println("Error fetching orders:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch orders"})
		return
	}

	var responseOrders []models.OrderResponse
	// Iterate through each order to fetch associated order items
	for _, order := range orders {
		var orderResponse models.OrderResponse
		// Map basic order fields to response structure
		orderResponse.ID = order.ID
		orderResponse.TotalPrice = order.TotalPrice
		orderResponse.FinalPrice = order.FinalPrice
		orderResponse.Status = order.Status

		// Fetch order items using GORM
		var items []models.ResponseOrderItem
		if err := db.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			log.Println("Error fetching order items:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch order items"})
			return
		}

		// Create a map to aggregate items by ItemID
		itemMap := make(map[int]models.ResponseOrderItem)

		// Iterate over order items and aggregate the data
		for _, item := range items {
			// Debugging output to check item data
			log.Printf("Processing item: %v", item)

			if existingItem, found := itemMap[item.ItemID]; found {
				// If the item already exists, update the quantity and price
				existingItem.Quantity += item.Quantity
				existingItem.Price += item.Price
				itemMap[item.ItemID] = existingItem
			} else {
				// If the item does not exist, add it to the map
				itemMap[item.ItemID] = models.ResponseOrderItem{
					ItemID:   item.ItemID,
					Quantity: item.Quantity,
					Price:    item.Price,
				}
			}
		}

		// Debugging output to check aggregated items
		log.Printf("Aggregated items for order %d: %v", order.ID, itemMap)

		// Convert the itemMap to a slice and add it to the response, converting to ItemResponse type
		for _, aggregatedItem := range itemMap {
			// Convert aggregatedItem (ResponseOrderItem) to ItemResponse
			orderResponse.Items = append(orderResponse.Items, models.ItemResponse{
				ItemID:   aggregatedItem.ItemID,
				Quantity: aggregatedItem.Quantity,
				Price:    aggregatedItem.Price,
			})
		}

		// Append the fully populated order response to the final response slice
		responseOrders = append(responseOrders, orderResponse)
	}

	// Return all orders with their aggregated items
	c.JSON(http.StatusOK, gin.H{
		"orders": responseOrders,
	})
}

// GetOrderByOrderId retrieves an order by its ID along with associated items using GORM
func GetOrderByOrderId(c *gin.Context, db *gorm.DB) {
	// Get the order ID from URL parameter
	id := c.Param("id")

	// Fetch the order by ID, including soft-delete check
	var order models.Order
	if err := db.Where("id = ? AND deleted_at IS NULL", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			log.Println("Error fetching order:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch order data"})
		}
		return
	}

	// Fetch the order items for the specific order
	var items []models.ResponseOrderItem
	if err := db.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		log.Println("Error fetching order items:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch order items"})
		return
	}

	// Prepare the response structure for the order
	responseOrder := models.OrderResposnse{
		ID:         order.ID,
		UserID:     order.UserID,
		TotalPrice: order.TotalPrice,
		FinalPrice: order.FinalPrice,
		Status:     order.Status,
		Items:      items,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		DeletedAt:  order.DeletedAt,
	}

	// Return the order with its items
	c.JSON(http.StatusOK, responseOrder)
}

// UpdateOrderByOrderId updates an order and its associated items
func UpdateOrderByOrderId(c *gin.Context, db *gorm.DB) {
	// Ensure the Gin logger is in Debug Mode
	gin.SetMode(gin.DebugMode)

	// Get the order ID from URL parameter
	idStr := c.Param("id")

	// Convert the string ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Bind the incoming JSON data to the updatedOrder struct
	var updatedOrder models.Order
	if err := c.ShouldBindJSON(&updatedOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Start a GORM transaction to ensure the order and order items are updated properly
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback() // Ensure rollback in case of error

	// Fetch the existing order
	var existingOrder models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&existingOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		log.Println("Error fetching order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Update the order status if it has changed
	if updatedOrder.Status != existingOrder.Status {
		if err := tx.Model(&existingOrder).Update("status", updatedOrder.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}
	}

	// Delete all existing items for this order
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old order items"})
		return
	}

	// Process each item in the updatedOrder.Items
	for _, updatedItem := range updatedOrder.Items {
		var itemPrice float64

		// Fetch the price of the item from the database
		if err := tx.Model(&models.Item{}).Where("id = ?", updatedItem.ItemID).Pluck("price", &itemPrice).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid item ID: %d", updatedItem.ItemID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item price"})
			return
		}

		// Insert the new item
		orderItem := models.OrderItem{
			OrderID:  id, // Cast to uint since GORM expects an unsigned integer
			ItemID:   updatedItem.ItemID,
			Quantity: updatedItem.Quantity,
			Price:    itemPrice,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert order item"})
			return
		}
	}

	// Recalculate the total price for the order
	var totalPrice float64
	if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", id).Select("SUM(price * quantity)").Scan(&totalPrice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate total price"})
		return
	}

	// Update the total price in the orders table
	if err := tx.Model(&models.Order{}).Where("id = ?", id).Update("total_price", totalPrice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order total price"})
		return
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Order and items updated successfully",
	})

	// Debug: Print success message
	fmt.Println("Transaction committed successfully.")
}

// UpdateOrderStatusByOrderId updates the order status to 'Confirm' if it is currently 'Pending'
func UpdateOrderStatusByOrderId(c *gin.Context, db *gorm.DB) {
	// Get the order ID from URL parameter
	idStr := c.Param("id")

	// Convert the string ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Start a GORM transaction to ensure the order status is updated properly
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback() // Ensure rollback in case of error

	// Fetch the current status of the order
	var order models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order status"})
		return
	}

	// Check if the order status is 'Pending'
	if order.Status != "Pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Order status is not 'Pending' (current status: %s)", order.Status)})
		return
	}

	// Update the status to 'Confirm'
	if err := tx.Model(&order).Update("status", "Confirm").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Order has been confirmed and placed successfully",
	})
}

// DeleteOrderByOrderId deletes an order (marks it as deleted) and updates its status to "cancelled"
func DeleteOrderByOrderId(c *gin.Context, db *gorm.DB) {
	// Get the order ID from URL parameter
	idStr := c.Param("id")

	// Convert the string ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Start a GORM transaction to ensure the order is deleted properly
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback() // Ensure rollback in case of error

	// Fetch the existing order to check if it is already deleted
	var order models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// If the order is already deleted, return an error
	if order.DeletedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order already deleted"})
		return
	}

	// Mark the order as deleted and update its status to "Cancelled"
	order.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	order.Status = "Cancelled"
	if err := tx.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Order deleted and status set to 'Cancelled' successfully",
	})
}

func calculateDiscounts(db *gorm.DB, order models.Order, items []models.OrderItem) models.Discounts {
	discounts := models.Discounts{}

	// Seasonal discount (e.g., December 3 - December 31)
	currentDate := time.Now()
	if currentDate.Month() == time.December && currentDate.Day() >= 3 && currentDate.Day() <= 31 {
		discounts.SeasonalDiscount = 0.15
		log.Println("Seasonal discount applied: 15%")
	}

	// Volume-based discount (10 or more units of any single item)
	for _, item := range items {
		if item.Quantity >= 10 {
			volumeDiscount := 0.10 * item.Price * float64(item.Quantity)
			discounts.VolumeBasedDiscount += volumeDiscount
			log.Printf("Volume discount for item %d: %.2f", item.ItemID, volumeDiscount)
		}
	}

	// Loyalty discount (if the user has more than 5 orders)
	var orderCount int64
	// Use GORM to count the number of orders for the user
	err := db.Model(&models.Order{}).Where("user_id = ?", order.UserID).Count(&orderCount).Error
	if err != nil {
		log.Printf("Error fetching user order count: %v", err)
	}

	if orderCount >= 5 {
		discounts.LoyaltyDiscount = 0.05
		log.Println("Loyalty discount applied: 5%")
	}

	return discounts
}

func calculateTotalPrice(db *gorm.DB, items []models.OrderItem, discounts models.Discounts) float64 {
	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	seasonalDiscount := totalPrice * discounts.SeasonalDiscount
	loyaltyDiscount := totalPrice * discounts.LoyaltyDiscount
	volumeDiscount := discounts.VolumeBasedDiscount

	totalDiscount := seasonalDiscount + loyaltyDiscount + volumeDiscount
	if totalDiscount > totalPrice {
		totalDiscount = totalPrice
	}

	finalPrice := totalPrice - totalDiscount
	log.Printf("Total Price: %.2f, Discounts: %.2f, Final Price: %.2f", totalPrice, totalDiscount, finalPrice)
	return finalPrice
}
