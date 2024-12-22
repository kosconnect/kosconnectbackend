package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/organisasi/kosconnectbackend/controllers"
	"github.com/organisasi/kosconnectbackend/middlewares"
)

func AuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", controllers.Register)
		authGroup.POST("/login", controllers.Login)
	}
}

func UserRoutes(router *gin.Engine) {
	api := router.Group("/api/users")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateUser)
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllUsers)
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetUserByID)
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateUser)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteUser)
	}
}
