package main

import (
	"log"
	"os"

	"reploy/internal/auth"
	"reploy/internal/handlers"
	"reploy/internal/middleware"
	"reploy/internal/models"
	"reploy/internal/storage"
	"reploy/internal/studlist"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	db, err := storage.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	oauthCfg := auth.NewOAuthConfig(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URL"),
	)

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public pages
	r.GET("/", func(c *gin.Context) {
		// Redirect to dashboard if already logged in
		if _, err := c.Cookie("token"); err == nil {
			c.Redirect(302, "/dashboard")
			return
		}
		c.HTML(200, "index.html", nil)
	})

	// Auth routes
	r.GET("/auth/google", handlers.GoogleLogin(oauthCfg))
	r.GET("/auth/google/callback", handlers.GoogleCallback(db, oauthCfg))
	r.POST("/auth/logout", handlers.Logout())

	// Student dashboard & editor pages (HTML)
	authed := r.Group("/")
	authed.Use(middleware.AuthRequired())
	{
		authed.GET("/dashboard", func(c *gin.Context) {
			c.HTML(200, "dashboard.html", gin.H{
				"pyccode": c.GetString("pyccode"),
				"role":    c.GetString("role"),
			})
		})
		authed.GET("/editor/:id", func(c *gin.Context) {
			c.HTML(200, "editor.html", gin.H{
				"app_id":  c.Param("id"),
				"pyccode": c.GetString("pyccode"),
			})
		})
		authed.GET("/shelf", func(c *gin.Context) {
			c.HTML(200, "shelf.html", nil)
		})
		authed.GET("/admin", middleware.TeacherRequired(), func(c *gin.Context) {
			c.HTML(200, "admin.html", nil)
		})
	}

	// API routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/apps", handlers.ListApps(db))
		api.POST("/apps", handlers.CreateApp(db))
		api.GET("/apps/:id", handlers.GetApp(db))
		api.PUT("/apps/:id", handlers.UpdateApp(db))
		api.DELETE("/apps/:id", handlers.DeleteApp(db))
		api.GET("/shelf", func(c *gin.Context) {
			apps, err := db.ListApprovedApps(c.Request.Context())
			if err != nil {
				c.JSON(500, gin.H{"error": "Could not fetch shelf"})
				return
			}
			if apps == nil {
				apps = []*models.App{}
			}
			c.JSON(200, apps)
		})

		// Current user info
		api.GET("/me", func(c *gin.Context) {
			pyccode := c.GetString("pyccode")
			c.JSON(200, gin.H{
				"user_id":      c.GetString("user_id"),
				"pyccode":      pyccode,
				"role":         c.GetString("role"),
				"display_name": studlist.DisplayName(pyccode),
			})
		})
	}

	// Admin API routes
	admin := r.Group("/admin/api")
	admin.Use(middleware.AuthRequired(), middleware.TeacherRequired())
	{
		admin.GET("/users", handlers.AdminListUsers(db))
		admin.PUT("/users/:id/ban", handlers.AdminBanUser(db))
		admin.PUT("/users/:id/role", handlers.AdminSetRole(db))
		admin.GET("/apps", handlers.AdminListApps(db))
		admin.GET("/apps/:id", handlers.AdminGetApp(db))
		admin.PUT("/apps/:id/hide", handlers.AdminHideApp(db))
		admin.PUT("/apps/:id/approve", handlers.AdminSetApproved(db))
		admin.PUT("/apps/:id/content", handlers.UpdateApp(db))
	}

	// Public app preview — must be last to avoid route conflicts
	r.GET("/:pyccode/:slug", handlers.PreviewApp(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Reploy running on :%s", port)
	r.Run(":" + port)
}
