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

func BoardingHouseRoutes(router *gin.Engine) {
	api := router.Group("/api/boardinghouses")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateBoardingHouse)
		// api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllBoardingHouses)
		// api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetBoardingHouseByID)
		// api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateBoardingHouse)
		// api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteBoardingHouse)
	}
}

func FacilityType(router *gin.Engine) {
	api := router.Group("/api/facilitytypes")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateFacilityType)
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllFacilityTypes)
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetFacilityTypeByID)
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateFacilityType)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteFacilityType)
	}
}

func RoomFacility(router *gin.Engine) {
	api := router.Group("/api/roomfacilities")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateRoomFacility)
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllRoomFacilities)
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetRoomFacilityByID)
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateRoomFacility)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteRoomFacility)
	}
}