package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	"gorm.io/gorm"
)

// var db *gorm.DB // Global variable to hold the database connection
//

// addUser creates a new user
// AddUser adds a new user to the database using GORM
func AddUser(c *gin.Context, db *gorm.DB) {

	if db == nil {
		log.Println("Database connection is nil!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	var newUser models.User
	// Bind incoming JSON data to the User struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Println("Error binding JSON:", err)
		return
	}

	// Log the user data to ensure it's being received correctly
	log.Println("Received user data:", newUser)

	// Insert the new user into the database using GORM
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user", "details": err.Error()})
		log.Println("Error inserting user:", err)
		return
	}
	var resUser models.UserResponse
	resUser.ID = newUser.ID
	resUser.Name = newUser.Name
	resUser.Email = newUser.Email
	resUser.CreatedAt = newUser.CreatedAt
	resUser.UpdatedAt = newUser.UpdatedAt
	resUser.DeletedAt = newUser.DeletedAt

	// Log the successful insertion
	log.Println("User added successfully:", newUser)

	// Return the response with the new user details
	c.JSON(http.StatusOK, gin.H{
		"message": "User added successfully",
		"user":    resUser,
	})
}

// FetchUsers fetches all users that are not deleted from the database using GORM
func FetchUsers(c *gin.Context, db *gorm.DB) {
	// Define a slice to hold the response users
	var users []models.User

	// Query to fetch only non-deleted users using GORM's Where method
	if err := db.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		// Log the error and respond with an internal server error
		log.Println("Error fetching users:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch users", "details": err.Error()})
		return
	}

	// Map users to a simplified response structure
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// Return the list of users as the response
	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
	})
}

// GetUserDetailByUserId fetches the details of a user by their ID
func GetUserDetailByUserId(c *gin.Context, db *gorm.DB) {
	// Get the user ID from the URL parameter
	id := c.Param("id")

	var user models.User
	// Query the database to fetch the user details using GORM
	if err := db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		// Handle case where user does not exist or any other error
		if err == gorm.ErrRecordNotFound {
			// User not found in the database
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "User not found",
			})
		} else {
			// Internal server error occurred while querying the database
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Unable to fetch user data",
				"details": err.Error(),
			})
		}
		return
	}

	var userResponse models.UserResponse
	userResponse.ID = user.ID
	userResponse.Name = user.Name
	userResponse.Email = user.Email
	userResponse.CreatedAt = user.CreatedAt
	userResponse.UpdatedAt = user.UpdatedAt
	userResponse.DeletedAt = user.DeletedAt

	// Successfully found the user and it is not soft-deleted
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User found",
		"user":    userResponse,
	})
}

// getUserDetailsWithOrders retrieves a user by their ID along with their orders and order items.
func GetUserDetailsWithOrdersByUserId(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	// Fetch user details and associated orders in one query using Preload
	var user models.User
	if err := db.Preload("Orders.Items").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch user data"})
		log.Println("Error fetching user:", err)
		return
	}

	// Check if the user is soft-deleted
	if user.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"message": "User has been soft deleted"})
		return
	}

	// Map orders and their items to response structs
	ordersResponse := make([]models.OrderResponse, len(user.Orders))
	for i, order := range user.Orders {
		orderResponse := models.OrderResponse{
			ID:         order.ID,
			TotalPrice: order.TotalPrice,
			FinalPrice: order.FinalPrice,
			Status:     order.Status,
		}

		// Map items to response struct
		itemsResponse := make([]models.ItemResponse, len(order.Items))
		for j, item := range order.Items {
			itemsResponse[j] = models.ItemResponse{
				ItemID:   item.ID,
				Price:    item.Price,
				Quantity: item.Quantity,
			}
		}
		orderResponse.Items = itemsResponse
		ordersResponse[i] = orderResponse
	}

	// Construct the full response combining user details and orders
	userResponse := models.UserOrderResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		DeletedAt:     user.DeletedAt,
		OrderResponse: ordersResponse,
	}

	// Send the final response with user and order details
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse,
	})
}

// UpdateUserDetails updates the user's details in the database using GORM
func UpdateUserDetails(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Find the user by ID
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		log.Println("Error fetching user:", err)
		return
	}

	// Update user details
	user.Name = updatedUser.Name
	user.Email = updatedUser.Email
	user.UpdatedAt = updatedUser.UpdatedAt // assuming updatedAt is passed as part of the request
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
	})
}

// DeleteUserByUserId deletes a user (soft delete) using GORM
func DeleteUserByUserId(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	// Find the user by ID
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		log.Println("Error fetching user:", err)
		return
	}

	// Check if the user is already soft deleted
	if user.DeletedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already deleted"})
		return
	}

	// Proceed with soft delete (set deleted_at field)
	if err := db.Model(&user).Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
