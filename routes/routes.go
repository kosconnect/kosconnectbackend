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
		// Hanya "owner" yang bisa membuat custom facility
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateCustomFacility)

		// Semua pengguna bisa mengambil semua fasilitas
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllCustomFacilities)

		// Hanya "owner" atau pengguna dengan akses tertentu yang bisa mengambil fasilitas berdasarkan ID
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetCustomFacilityByID)

		// Hanya "owner" yang bisa mengupdate atau menghapus custom facility
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateCustomFacility)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteCustomFacility)

		// Rute untuk mengambil fasilitas khusus berdasarkan owner ID
		api.GET("/owner", middlewares.JWTAuthMiddleware(), controllers.GetCustomFacilitiesByOwnerID)
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
	api := router.Group("/api/boardingHouses")
	{
		// Public route
		api.GET("/", controllers.GetAllBoardingHouse)

		// Protected routes - Requires JWT authentication
		api.Use(middlewares.JWTAuthMiddleware())
		{
			api.POST("/", controllers.CreateBoardingHouse)
			api.GET("/owner", controllers.GetBoardingHouseByOwnerID)
			api.PUT("/:id", controllers.UpdateBoardingHouse)
			api.DELETE("/:id", controllers.DeleteBoardingHouse)
		}
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
