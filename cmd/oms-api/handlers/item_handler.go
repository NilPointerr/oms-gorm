package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	"gorm.io/gorm"
)

// var db *sql.DB // Global variable to hold the database connection
//

func AddItem(c *gin.Context, db *gorm.DB) {
	var newItem models.Item
	// Bind incoming JSON data to the Item struct
	if err := c.ShouldBindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate the fields (check for empty strings or invalid price)
	if newItem.Name == "" || newItem.Description == "" || newItem.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields must be filled and price must be positive"})
		return
	}

	// Log the item data to ensure it's being received correctly
	fmt.Println(newItem.Name, newItem.Description, newItem.Price)

	// Use GORM to insert the new item into the database
	if err := db.Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert item"})
		return
	}

	// Return the response with the new item details
	c.JSON(http.StatusOK, gin.H{
		"message": "Item added successfully",
		"item":    newItem,
	})
}

// GetItems responds with the list of all items as JSON
func GetItems(c *gin.Context, db *gorm.DB) {
	var items []models.Item

	// Fetch all non-deleted items from the database
	if err := db.Where("deleted_at IS NULL").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch data"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetItemByItemId retrieves a single item by its ID
func GetItemByItemId(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var item models.Item

	// Fetch the item by ID
	if err := db.Where("id = ? AND deleted_at IS NULL", id).First(&item).Error; err != nil {
		// Handle different error cases
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch data"})
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateItem updates an existing item
func UpdateItemByItemId(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var updatedItem models.Item
	if err := c.ShouldBindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Find the item by ID
	var item models.Item
	if err := db.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Update the item's fields
	item.Name = updatedItem.Name
	item.Description = updatedItem.Description
	item.Price = updatedItem.Price
	item.UpdatedAt = time.Now() // Ensure UpdatedAt is set to the current time

	// Save the updated item
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated successfully",
	})
}

// DeleteItem deletes an item
func DeleteItemByItemId(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	// Find the item by ID
	var item models.Item
	if err := db.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Check if the item is already soft-deleted
	if item.DeletedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item is already deleted"})
		return
	}

	// Proceed with soft delete (setting deleted_at to the current time)
	item.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
	})
}
