package handler

import (
	"log"

	"QuanPhotos/internal/config"
	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/jwt"
	"QuanPhotos/internal/pkg/storage"
	"QuanPhotos/internal/repository/postgresql/category"
	"QuanPhotos/internal/repository/postgresql/comment"
	"QuanPhotos/internal/repository/postgresql/photo"
	"QuanPhotos/internal/repository/postgresql/ranking"
	"QuanPhotos/internal/repository/postgresql/share"
	"QuanPhotos/internal/repository/postgresql/tag"
	"QuanPhotos/internal/repository/postgresql/ticket"
	"QuanPhotos/internal/repository/postgresql/token"
	"QuanPhotos/internal/repository/postgresql/user"
	adminService "QuanPhotos/internal/service/admin"
	"QuanPhotos/internal/service/auth"
	categoryService "QuanPhotos/internal/service/category"
	commentService "QuanPhotos/internal/service/comment"
	photoService "QuanPhotos/internal/service/photo"
	rankingService "QuanPhotos/internal/service/ranking"
	shareService "QuanPhotos/internal/service/share"
	"QuanPhotos/internal/service/system"
	tagService "QuanPhotos/internal/service/tag"
	ticketService "QuanPhotos/internal/service/ticket"
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
	systemHandler   *SystemHandler
	authHandler     *AuthHandler
	userHandler     *UserHandler
	adminHandler    *AdminHandler
	photoHandler    *PhotoHandler
	ticketHandler   *TicketHandler
	categoryHandler *CategoryHandler
	tagHandler      *TagHandler
	commentHandler  *CommentHandler
	shareHandler    *ShareHandler
	publicHandler   *PublicHandler
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
	ticketRepo := ticket.NewTicketRepository(db)
	categoryRepo := category.NewCategoryRepository(db)
	tagRepo := tag.NewTagRepository(db)
	commentRepo := comment.NewCommentRepository(db)
	shareRepo := share.NewShareRepository(db)
	rankingRepo := ranking.NewRankingRepository(db)

	// Initialize local storage
	localStorage, err := storage.NewLocalStorage(cfg.Storage.Path, cfg.Storage.BaseURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize local storage: %v", err)
	}

	// Initialize services
	systemService := system.NewService(cfg)
	authService := auth.New(db, userRepo, tokenRepo, jwtManager)
	userSvc := userService.New(userRepo)
	adminSvc := adminService.NewFull(userRepo, photoRepo, ticketRepo, cfg.Storage.BaseURL)

	// Initialize photo service with uploader if storage is available
	var photoSvc *photoService.Service
	if localStorage != nil {
		photoSvc = photoService.NewWithUploader(photoRepo, localStorage, cfg)
	} else {
		photoSvc = photoService.New(photoRepo, cfg.Storage.BaseURL)
	}

	// Initialize ticket service
	ticketSvc := ticketService.New(ticketRepo, cfg.Storage.BaseURL)

	// Initialize category and tag services
	categorySvc := categoryService.New(categoryRepo, cfg.Storage.BaseURL)
	tagSvc := tagService.New(tagRepo, cfg.Storage.BaseURL)

	// Initialize comment, share, and ranking services
	commentSvc := commentService.New(commentRepo, cfg.Storage.BaseURL)
	shareSvc := shareService.New(shareRepo, cfg.Storage.BaseURL)
	rankingSvc := rankingService.New(rankingRepo, cfg.Storage.BaseURL)

	// Initialize handlers
	systemHandler := NewSystemHandler(systemService)
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userSvc)
	adminHandler := NewAdminHandler(adminSvc)
	photoHandler := NewPhotoHandler(photoSvc, cfg.Storage.MaxSize)
	ticketHandler := NewTicketHandler(ticketSvc)
	categoryHandler := NewCategoryHandler(categorySvc)
	tagHandler := NewTagHandler(tagSvc)
	commentHandler := NewCommentHandler(commentSvc)
	shareHandler := NewShareHandler(shareSvc)
	publicHandler := NewPublicHandler(photoRepo, rankingSvc, cfg.Storage.BaseURL)

	return &Router{
		engine:          engine,
		config:          cfg,
		jwtManager:      jwtManager,
		systemHandler:   systemHandler,
		authHandler:     authHandler,
		userHandler:     userHandler,
		adminHandler:    adminHandler,
		photoHandler:    photoHandler,
		ticketHandler:   ticketHandler,
		categoryHandler: categoryHandler,
		tagHandler:      tagHandler,
		commentHandler:  commentHandler,
		shareHandler:    shareHandler,
		publicHandler:   publicHandler,
	}
}

// Setup sets up all routes
func (r *Router) Setup() {
	// Health check endpoint
	r.engine.GET("/health", r.systemHandler.Health)

	// Static file serving for uploads
	r.engine.Static("/uploads", r.config.Storage.Path)

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
			photos.GET("/:id/comments", middleware.OptionalAuth(r.jwtManager), r.commentHandler.List)

			// Protected routes (require authentication)
			photos.POST("", middleware.Auth(r.jwtManager), r.photoHandler.Upload)
			photos.GET("/mine", middleware.Auth(r.jwtManager), r.photoHandler.ListMine)
			photos.GET("/favorites", middleware.Auth(r.jwtManager), r.photoHandler.ListFavorites)
			photos.POST("/:id/favorite", middleware.Auth(r.jwtManager), r.photoHandler.AddFavorite)
			photos.DELETE("/:id/favorite", middleware.Auth(r.jwtManager), r.photoHandler.RemoveFavorite)
			photos.POST("/:id/like", middleware.Auth(r.jwtManager), r.photoHandler.AddLike)
			photos.DELETE("/:id/like", middleware.Auth(r.jwtManager), r.photoHandler.RemoveLike)
			photos.DELETE("/:id", middleware.Auth(r.jwtManager), r.photoHandler.Delete)
			photos.POST("/:id/comments", middleware.Auth(r.jwtManager), r.commentHandler.Create)
			photos.POST("/:id/share", middleware.Auth(r.jwtManager), r.shareHandler.Share)
		}

		// Comments routes (for individual comment operations)
		comments := v1.Group("/comments")
		comments.Use(middleware.Auth(r.jwtManager))
		{
			comments.DELETE("/:id", r.commentHandler.Delete)
			comments.POST("/:id/like", r.commentHandler.AddLike)
			comments.DELETE("/:id/like", r.commentHandler.RemoveLike)
		}

		// Categories routes
		categories := v1.Group("/categories")
		{
			// Public routes
			categories.GET("", r.categoryHandler.List)
			categories.GET("/:id", r.categoryHandler.GetByID)

			// Admin routes
			categories.POST("", middleware.Auth(r.jwtManager), middleware.RequireMinRole(model.RoleAdmin), r.categoryHandler.Create)
			categories.PUT("/:id", middleware.Auth(r.jwtManager), middleware.RequireMinRole(model.RoleAdmin), r.categoryHandler.Update)
			categories.DELETE("/:id", middleware.Auth(r.jwtManager), middleware.RequireMinRole(model.RoleAdmin), r.categoryHandler.Delete)
		}

		// Tags routes
		tags := v1.Group("/tags")
		{
			// Public routes
			tags.GET("", r.tagHandler.List)
			tags.GET("/search", r.tagHandler.Search)
			tags.GET("/:id/photos", r.tagHandler.ListPhotos)
		}

		// Featured photos routes (public)
		v1.GET("/featured", r.publicHandler.ListFeatured)

		// Rankings routes (public)
		rankings := v1.Group("/rankings")
		{
			rankings.GET("/photos", r.publicHandler.PhotoRanking)
			rankings.GET("/users", r.publicHandler.UserRanking)
		}

		// Public announcements routes
		announcements := v1.Group("/announcements")
		{
			announcements.GET("", r.publicHandler.ListAnnouncements)
			announcements.GET("/:id", r.publicHandler.GetAnnouncement)
		}

		// Tickets routes (require authentication)
		tickets := v1.Group("/tickets")
		tickets.Use(middleware.Auth(r.jwtManager))
		{
			tickets.POST("", r.ticketHandler.Create)
			tickets.GET("", r.ticketHandler.List)
			tickets.GET("/:id", r.ticketHandler.GetDetail)
			tickets.POST("/:id/replies", r.ticketHandler.Reply)
		}

		// Admin routes (require admin or superadmin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.Auth(r.jwtManager))
		admin.Use(middleware.RequireMinRole(model.RoleAdmin))
		{
			// User management
			admin.GET("/users", r.adminHandler.ListUsers)
			admin.PUT("/users/:id/role", r.adminHandler.UpdateUserRole)
			admin.PUT("/users/:id/status", r.adminHandler.UpdateUserStatus)

			// Photo reviews
			admin.GET("/reviews", r.adminHandler.ListReviews)
			admin.POST("/reviews/:id", r.adminHandler.ReviewPhoto)

			// Photo management
			admin.DELETE("/photos/:id", r.adminHandler.AdminDeletePhoto)

			// Ticket management
			admin.GET("/tickets", r.adminHandler.ListTickets)
			admin.PUT("/tickets/:id", r.adminHandler.ProcessTicket)

			// Featured photos
			admin.POST("/featured", r.adminHandler.AddFeatured)
			admin.DELETE("/featured/:id", r.adminHandler.RemoveFeatured)

			// Announcements
			admin.GET("/announcements", r.adminHandler.ListAnnouncements)
			admin.GET("/announcements/:id", r.adminHandler.GetAnnouncement)
			admin.POST("/announcements", r.adminHandler.CreateAnnouncement)
			admin.PUT("/announcements/:id", r.adminHandler.UpdateAnnouncement)
			admin.DELETE("/announcements/:id", r.adminHandler.DeleteAnnouncement)
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
