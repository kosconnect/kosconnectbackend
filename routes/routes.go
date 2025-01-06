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

		// Tambahkan routes untuk OAuth Google
		authGroup.GET("/google/login", controllers.HandleGoogleLogin)
		authGroup.GET("/callback", controllers.HandleGoogleCallback)
		authGroup.PUT("/assign-role", controllers.AssignRole)
	}
}

func UserRoutes(router *gin.Engine) {
    api := router.Group("/api/users")
    {
        api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateUser)                        // Admin creates a user
        api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllUsers)                       // Admin views all users
        api.GET("/me", middlewares.JWTAuthMiddleware(), controllers.GetMyAccount)                   // Logged-in user views their own account
        api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetUserByID)                   // Get user by ID
        api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateUser)                    // Update user details
        api.PUT("/:id/role", middlewares.JWTAuthMiddleware(), controllers.UpdateUserRole)           // Admin updates user role
		api.PUT("/change-password", middlewares.JWTAuthMiddleware(), controllers.ChangePassword)    // Logged-in user changes their password
        api.PUT("/:id/reset-password", middlewares.JWTAuthMiddleware(), controllers.ResetPassword)  // Admin resets user password
        api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteUser)                 // Delete a user
		// Tambahkan route untuk SetUserRole dan CheckUserRole
		// api.PUT("/role", middlewares.JWTAuthMiddleware(), controllers.SetUserRole) // Set user role (user/owner)
		// api.GET("/role", middlewares.JWTAuthMiddleware(), controllers.CheckUserRole) // Check user role
    }
}


func CustomFacility(router *gin.Engine) {
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

func BoardingHouse(router *gin.Engine) {
	api := router.Group("/api/boardingHouses")
	{
		// Public route
		api.GET("/", controllers.GetAllBoardingHouse)

		// Protected routes - Requires JWT authentication
		api.Use(middlewares.JWTAuthMiddleware())
		{
			api.POST("/", controllers.CreateBoardingHouse)
			api.GET("/owner", controllers.GetBoardingHouseByOwnerID)
			api.GET("/:id", controllers.GetBoardingHouseByID)
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

func RoomRoutes(router *gin.Engine) {
	// Group routes for room
	api := router.Group("/rooms")

	// Apply middleware for authorization (if needed)
	api.Use(middlewares.JWTAuthMiddleware())
	{
		// Public endpoint to get all rooms
		api.GET("/", controllers.GetAllRoom)

		// Public endpoint to get rooms by Boarding House ID
		api.GET("/boarding-house/:id", controllers.GetRoomByBoardingHouseID)

		// Public endpoint to get room by ID
		api.GET("/:id", controllers.GetRoomByID)

		// Protected endpoints for owners/admin to manage rooms
		api.POST("/", controllers.CreateRoom) // Create room
		api.PUT("/:id", controllers.UpdateRoom) // Update room
		api.DELETE("/:id", controllers.DeleteRoom) // Delete room
	}
}

