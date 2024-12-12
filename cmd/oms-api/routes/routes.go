package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/keyurKalariya/OMS/cmd/oms-api/handlers" // Import the handlers package
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {

	// Users API routes
	r.POST("/api/createUser", func(c *gin.Context) { handlers.AddUser(c, db) })
	r.GET("/api/FetchAllUser", func(c *gin.Context) { handlers.FetchUsers(c, db) })
	r.GET("/api/GetUserDetailByUserId/:id", func(c *gin.Context) { handlers.GetUserDetailByUserId(c, db) })
	r.GET("/api/GetUserDetailsWithOrdersByUserId/:id", func(c *gin.Context) { handlers.GetUserDetailsWithOrdersByUserId(c, db) })
	r.PUT("/api/UpdateUserDetails/:id", func(c *gin.Context) { handlers.UpdateUserDetails(c, db) })
	r.DELETE("/api/DeleteUserByUserId/:id", func(c *gin.Context) { handlers.DeleteUserByUserId(c, db) })

	//Items API routes
	r.POST("/api/AddItem", func(c *gin.Context) { handlers.AddItem(c, db) })
	r.GET("/api/GetItems", func(c *gin.Context) { handlers.GetItems(c, db) })
	r.GET("/api/GetItemByItemId/:id", func(c *gin.Context) { handlers.GetItemByItemId(c, db) })
	r.PUT("/api/UpdateItemByItemId/:id", func(c *gin.Context) { handlers.UpdateItemByItemId(c, db) })
	r.DELETE("/api/DeleteItemByItemId/:id", func(c *gin.Context) { handlers.DeleteItemByItemId(c, db) })

	//orders API routes
	r.POST("/api/createOrder", func(c *gin.Context) { handlers.CreateOrder(c, db) })
	r.GET("/api/getOrders", func(c *gin.Context) { handlers.GetOrders(c, db) })
	r.GET("/api/getOrderByOrderId/:id", func(c *gin.Context) { handlers.GetOrderByOrderId(c, db) })
	r.PUT("/api/updateOrderByOrderId/:id", func(c *gin.Context) { handlers.UpdateOrderByOrderId(c, db) })
	r.PUT("/api/updateOrderStatusByOrderId/:id", func(c *gin.Context) { handlers.UpdateOrderStatusByOrderId(c, db) })
	r.DELETE("/api/deleteOrderByOderId/:id", func(c *gin.Context) { handlers.DeleteOrderByOrderId(c, db) })

}
