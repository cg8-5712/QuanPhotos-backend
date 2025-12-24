package handler

import (
	"QuanPhotos/internal/config"
	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/jwt"
	"QuanPhotos/internal/repository/postgresql/photo"
	"QuanPhotos/internal/repository/postgresql/token"
	"QuanPhotos/internal/repository/postgresql/user"
	adminService "QuanPhotos/internal/service/admin"
	"QuanPhotos/internal/service/auth"
	photoService "QuanPhotos/internal/service/photo"
	"QuanPhotos/internal/service/system"
	userService "QuanPhotos/internal/service/user"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Router holds the gin router instance
type Router struct {
	engine *gin.Engine
	config *config.Config

	// JWT manager
	jwtManager *jwt.Manager

	// Handlers
	systemHandler *SystemHandler
	authHandler   *AuthHandler
	userHandler   *UserHandler
	adminHandler  *AdminHandler
	photoHandler  *PhotoHandler
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, db *sqlx.DB) *Router {
	// Set gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Apply global middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.RequestID())

	// Apply CORS if enabled
	if cfg.CORS.Enabled {
		engine.Use(middleware.CORS(middleware.CORSConfig{
			AllowedOrigins: cfg.CORS.AllowedOrigins,
			AllowedMethods: cfg.CORS.AllowedMethods,
			AllowedHeaders: cfg.CORS.AllowedHeaders,
			MaxAge:         cfg.CORS.MaxAge,
		}))
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
		cfg.JWT.Issuer,
	)

	// Initialize repositories
	userRepo := user.NewUserRepository(db)
	tokenRepo := token.NewTokenRepository(db)
	photoRepo := photo.NewPhotoRepository(db)

	// Initialize services
	systemService := system.NewService(cfg)
	authService := auth.New(db, userRepo, tokenRepo, jwtManager)
	userSvc := userService.New(userRepo)
	adminSvc := adminService.New(userRepo)
	photoSvc := photoService.New(photoRepo, cfg.Storage.BaseURL)

	// Initialize handlers
	systemHandler := NewSystemHandler(systemService)
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userSvc)
	adminHandler := NewAdminHandler(adminSvc)
	photoHandler := NewPhotoHandler(photoSvc)

	return &Router{
		engine:        engine,
		config:        cfg,
		jwtManager:    jwtManager,
		systemHandler: systemHandler,
		authHandler:   authHandler,
		userHandler:   userHandler,
		adminHandler:  adminHandler,
		photoHandler:  photoHandler,
	}
}

// Setup sets up all routes
func (r *Router) Setup() {
	// Health check endpoint
	r.engine.GET("/health", r.systemHandler.Health)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// System routes
		v1.GET("/system/info", r.systemHandler.Info)

		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", r.authHandler.Register)
			authRoutes.POST("/login", r.authHandler.Login)
			authRoutes.POST("/refresh", r.authHandler.Refresh)
			authRoutes.POST("/logout", r.authHandler.Logout)
		}

		// User routes
		users := v1.Group("/users")
		{
			// Public routes
			users.GET("/:id", r.userHandler.GetUser)
			users.GET("/:id/photos", r.photoHandler.ListUserPhotos)

			// Protected routes (require authentication)
			users.GET("/me", middleware.Auth(r.jwtManager), r.userHandler.GetCurrentUser)
			users.PUT("/me", middleware.Auth(r.jwtManager), r.userHandler.UpdateCurrentUser)
			users.PUT("/me/password", middleware.Auth(r.jwtManager), r.userHandler.ChangePassword)
		}

		// Photos routes
		photos := v1.Group("/photos")
		{
			// Public routes
			photos.GET("", r.photoHandler.List)
			photos.GET("/:id", middleware.OptionalAuth(r.jwtManager), r.photoHandler.GetDetail)

			// Protected routes (require authentication)
			photos.GET("/mine", middleware.Auth(r.jwtManager), r.photoHandler.ListMine)
			photos.GET("/favorites", middleware.Auth(r.jwtManager), r.photoHandler.ListFavorites)
			photos.POST("/:id/favorite", middleware.Auth(r.jwtManager), r.photoHandler.AddFavorite)
			photos.DELETE("/:id/favorite", middleware.Auth(r.jwtManager), r.photoHandler.RemoveFavorite)
			photos.POST("/:id/like", middleware.Auth(r.jwtManager), r.photoHandler.AddLike)
			photos.DELETE("/:id/like", middleware.Auth(r.jwtManager), r.photoHandler.RemoveLike)
			photos.DELETE("/:id", middleware.Auth(r.jwtManager), r.photoHandler.Delete)
		}

		// Categories routes (to be implemented)
		// categories := v1.Group("/categories")

		// Tags routes (to be implemented)
		// tags := v1.Group("/tags")

		// Tickets routes (to be implemented)
		// tickets := v1.Group("/tickets")

		// Admin routes (require admin or superadmin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.Auth(r.jwtManager))
		admin.Use(middleware.RequireMinRole(model.RoleAdmin))
		{
			// User management
			admin.GET("/users", r.adminHandler.ListUsers)
			admin.PUT("/users/:id/role", r.adminHandler.UpdateUserRole)
			admin.PUT("/users/:id/status", r.adminHandler.UpdateUserStatus)

			// Reviews routes (to be implemented)
			// admin.GET("/reviews", r.adminHandler.ListReviews)
			// admin.POST("/reviews/:id", r.adminHandler.ReviewPhoto)

			// Tickets routes (to be implemented)
			// admin.GET("/tickets", r.adminHandler.ListTickets)
			// admin.PUT("/tickets/:id", r.adminHandler.ProcessTicket)

			// Featured routes (to be implemented)
			// admin.POST("/featured", r.adminHandler.CreateFeatured)
			// admin.DELETE("/featured/:id", r.adminHandler.DeleteFeatured)

			// Announcements routes (to be implemented)
			// admin.GET("/announcements", r.adminHandler.ListAnnouncements)
			// admin.POST("/announcements", r.adminHandler.CreateAnnouncement)
			// admin.PUT("/announcements/:id", r.adminHandler.UpdateAnnouncement)
			// admin.DELETE("/announcements/:id", r.adminHandler.DeleteAnnouncement)

			// Photos routes (to be implemented)
			// admin.DELETE("/photos/:id", r.adminHandler.DeletePhoto)
		}

		// Superadmin routes (to be implemented)
		// superadmin := v1.Group("/superadmin")
		// superadmin.Use(middleware.Auth(r.jwtManager))
		// superadmin.Use(middleware.RequireSuperAdmin())
	}
}

// GetEngine returns the gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run() error {
	return r.engine.Run(":" + r.config.App.Port)
}
