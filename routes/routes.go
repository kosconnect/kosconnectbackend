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
		authGroup.GET("/verify", controllers.VerifyEmail)
		authGroup.POST("/login", controllers.Login)

		// Tambahkan routes untuk OAuth Google
		authGroup.GET("/google/login", controllers.HandleGoogleLogin)
		authGroup.GET("/callback", controllers.HandleGoogleCallback)
		authGroup.PUT("/assign-role", controllers.AssignRole)
		authGroup.POST("/googleauth", controllers.GoogleAuth)
	}
}

func UserRoutes(router *gin.Engine) {
	api := router.Group("/api/users")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateUser)                     // Admin creates a user
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllUsers)                     // Admin views all users
		api.GET("/owner", middlewares.JWTAuthMiddleware(), controllers.GetAllOwners)               // ambil semua data owner
		api.GET("/:id/owner", middlewares.JWTAuthMiddleware(), controllers.GetOwnerByID)               // ambil semua data owner
		api.GET("/me", middlewares.JWTAuthMiddleware(), controllers.GetMyAccount)                  // Logged-in user views their own account
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetUserByID)                  // Get user by ID
		api.PUT("/me", middlewares.JWTAuthMiddleware(), controllers.UpdateMe)                      // Update user details for user yg login
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateUser)                   // Update user details oleh admin
		api.PUT("/:id/role", middlewares.JWTAuthMiddleware(), controllers.UpdateUserRole)          // Admin updates user role
		api.PUT("/change-password", middlewares.JWTAuthMiddleware(), controllers.ChangePassword)   // berdasarkan pengguna yang login
		api.PUT("/:id/reset-password", middlewares.JWTAuthMiddleware(), controllers.ResetPassword) // Admin bisa reset password pengguna lain
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteUser)                // Delete a user
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

		// Rute untuk mengambil fasilitas khusus berdasarkan owner ID yang disimpan di query atau url disisi admin
		api.GET("/admin", middlewares.JWTAuthMiddleware(), controllers.GetCustomFacilitiesByOwnerIDAdmin)
	}
}

func CategoryRoutes(router *gin.Engine) {
	api := router.Group("/api/categories")
	{
		api.GET("/", controllers.GetAllCategories)
		api.GET("/:id", controllers.GetCategoryByID)

		api.Use(middlewares.JWTAuthMiddleware())
		{
			api.POST("/", controllers.CreateCategory)
			api.PUT("/:id", controllers.UpdateCategory)
			api.DELETE("/:id", controllers.DeleteCategory)
		}
	}
}

func BoardingHouse(router *gin.Engine) {
	api := router.Group("/api/boardingHouses")
	{
		// Public route
		api.GET("/", controllers.GetAllBoardingHouse)
		api.GET("/:id/detail", controllers.GetBoardingHouseDetails)
		api.GET("/:id", controllers.GetBoardingHouseByID)

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

func Facility(router *gin.Engine) {
	api := router.Group("/api/facility")
	{
		api.POST("/", middlewares.JWTAuthMiddleware(), controllers.CreateFacility)
		api.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetAllFacilities)
		// yg type ini buat get data fasilitas berdasarkan typenya, ada /api/facility/type?type=room dan /api/facility/type?type=boarding_house cara manggilnya
		api.GET("/type", middlewares.JWTAuthMiddleware(), controllers.GetFacilitiesByType)
		api.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetFacilityByID)
		api.PUT("/:id", middlewares.JWTAuthMiddleware(), controllers.UpdateFacility)
		api.DELETE("/:id", middlewares.JWTAuthMiddleware(), controllers.DeleteFacility)
	}
}

func RoomRoutes(router *gin.Engine) {
	// Group routes for room
	api := router.Group("/api/rooms")
	// Public endpoint to get room by ID

	api.GET("/:id/detail", controllers.GetRoomDetailsByID)
	api.GET("/:id/pages", controllers.GetRoomDetailPages)
	api.GET("/home", controllers.GetRoomsForLandingPage)
	// Public endpoint to get all rooms
	api.GET("/", controllers.GetAllRooms)

	// Apply middleware for authorization (if needed)
	api.Use(middlewares.JWTAuthMiddleware())
	{
		api.GET("/:id", controllers.GetRoomByID)
		// Public endpoint to get rooms by Boarding House ID
		api.GET("/boarding-house/:id", controllers.GetRoomByBoardingHouseID)

		// Protected endpoints for owners/admin to manage rooms
		api.POST("/:boardingHouseID", controllers.CreateRoom)
		api.PUT("/:id", controllers.UpdateRoom)    // Update room
		api.DELETE("/:id", controllers.DeleteRoom) // Delete room
	}
}

func TransactionRoutes(router *gin.Engine) {
	// Route untuk callback dari Midtrans (publik)
	router.POST("/midtrans/notification", controllers.PaymentNotification)

	// Route untuk membuat pembayaran
	router.POST("/transactions/:transaction_id/payment", controllers.CreatePayment)

	// Grup API dengan prefix /api/transaction
	api := router.Group("/api/transaction")
	api.Use(middlewares.JWTAuthMiddleware()) // Semua route dalam grup menggunakan middleware JWT
	{
		// Membuat transaksi baru
		api.POST("/", controllers.CreateTransaction)

		// Mendapatkan semua transaksi (Admin dan Owner)
		api.GET("/", controllers.GetAllTransactions)

		// Mendapatkan detail transaksi berdasarkan ID
		api.GET("/:id", controllers.GetTransactionByID)

		// Mendapatkan transaksi milik pengguna tertentu (User)
		api.GET("/user", controllers.GetTransactionsByUser)
		api.GET("/admin/user/:id", controllers.GetTransactionsUserByAdmin)

		// Mendapatkan transaksi milik owner tertentu (Owner)
		api.GET("/owner", controllers.GetTransactionsByOwner)
		api.GET("/admin/owner/:id", controllers.GetTransactionsOwnerByAdmin)

		// Mendapatkan transaksi berdasarkan status pembayaran (Pending, Paid, etc.)
		api.GET("/status/:status", controllers.GetTransactionsByPaymentStatus)

		// Memperbarui status pembayaran transaksi (misalnya: Paid, Cancelled, dll.)
		api.PUT("/:id/payment-status", controllers.UpdateTransaction)

		// Menghapus transaksi (opsional, hanya untuk admin)
		api.DELETE("/:id", controllers.DeleteTransaction)
	}
}

