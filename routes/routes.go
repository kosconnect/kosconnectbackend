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

func CustomFacilityRoutes(router *gin.Engine) {
    api := router.Group("/api/customFacilities")
    {
        api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateCustomFacility)
        api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllCustomFacilities)
        api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetCustomFacilityByID)
        api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateCustomFacility)
        api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteCustomFacility)
    }
}

func CategoryRoutes(router *gin.Engine) {
	api := router.Group("/api/categories")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateCategory)
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllCategories)
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetCategoryByID)
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateCategory)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteCategory)
	}
}