package handlers

import (
	"net/http"

	"reploy/internal/storage"

	"github.com/gin-gonic/gin"
)

func PreviewApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pyccode := c.Param("pyccode")
		slug := c.Param("slug")

		app, err := db.GetAppByPYCCodeAndSlug(c.Request.Context(), pyccode, slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"message": "找不到此應用程式"})
			return
		}

		// Check if owner is banned
		owner, err := db.GetUserByID(c.Request.Context(), app.UserID)
		if err != nil || owner.IsBanned {
			c.HTML(http.StatusForbidden, "error.html", gin.H{"message": "此應用程式無法存取"})
			return
		}

		if app.IsHidden || !app.IsPublic {
			c.HTML(http.StatusForbidden, "error.html", gin.H{"message": "此應用程式已被下架"})
			return
		}

		// Serve the raw HTML with security headers
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", "connect-src 'none'")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(app.HTMLContent))
	}
}
