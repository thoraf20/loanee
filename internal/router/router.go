package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/thoraf20/loanee/internal/container"
	"github.com/thoraf20/loanee/internal/middleware"
)

// Setup configures all routes with handlers from container
func Setup(c *container.Container) *gin.Engine {
	// Set Gin mode based on environment
	if c.Config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware (order matters!)
	r.Use(middleware.Recovery(c.Logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(c.Logger))
	r.Use(middleware.CORS(c.Config))
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes - Auth
		auth := v1.Group("/auth")

		{
			auth.POST("/register", c.AuthHandler.Register)
			auth.POST("/login", c.AuthHandler.Login)
			auth.POST("/verify-email", c.AuthHandler.VerifyEmail)
			auth.POST("/resend-code", c.AuthHandler.ResendVerificationCode)
			auth.POST("/forgot-password", c.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", c.AuthHandler.ResetPassword)
			// auth.POST("/refresh", c.AuthHandler.RefreshToken)
		}

		// Protected routes - require authentication
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(c.JWTManager, c.TokenBlacklist))

		{
			protected.POST("/auth/logout", c.AuthHandler.Logout)

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", c.UserHandler.GetProfile)
				// users.PUT("/me", c.UserHandler.UpdateProfile)
			}

			collaterals := protected.Group("/collaterals")
			{
				collaterals.GET("", c.CollateralHandler.ListMine)
				collaterals.GET("/preview", c.CollateralHandler.Preview)
				collaterals.POST("/request", c.CollateralHandler.CreateRequest)
				collaterals.POST("/lock", c.CollateralHandler.Lock)
				collaterals.POST("/:id/release-request", c.CollateralHandler.RequestRelease)
			}

			wallets := protected.Group("/wallets")
			{
				wallets.GET("", c.WalletHandler.ListMine)
			}

			loans := protected.Group("/loans")
			{
				loans.GET("", c.LoanHandler.ListMine)
				loans.POST("/:id/repay", c.PaymentHandler.RepayLoan)
				loans.GET("/:id/repayments", c.PaymentHandler.ListRepayments)
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthRequired(c.JWTManager, c.TokenBlacklist))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/collaterals", c.CollateralHandler.AdminList)
			admin.PUT("/collaterals/:id/approve-release", c.CollateralHandler.AdminApproveRelease)
			admin.PUT("/collaterals/:id/reject-release", c.CollateralHandler.AdminRejectRelease)
			admin.GET("/loans", c.LoanHandler.AdminList)
			admin.PUT("/loans/:id/approve", c.LoanHandler.AdminApprove)
			admin.POST("/loans/:id/disburse", c.LoanHandler.AdminDisburse)
		}
	}

	return r
}
