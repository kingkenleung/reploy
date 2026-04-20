package handlers

import (
	"net/http"

	"reploy/internal/models"
	"reploy/internal/storage"

	"github.com/gin-gonic/gin"
)

const superAdmin = "lkh1@school.pyc.edu.hk"

func AdminListUsers(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := db.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
			return
		}
		if users == nil {
			users = []*models.User{}
		}
		c.JSON(http.StatusOK, users)
	}
}

func AdminBanUser(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Banned bool `json:"banned"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := db.SetUserBanned(c.Request.Context(), c.Param("id"), req.Banned); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
			return
		}
		msg := "User unbanned"
		if req.Banned {
			msg = "User banned"
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}

// AdminSetRole promotes or demotes a user. Only the superadmin may call this.
func AdminSetRole(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify caller is the superadmin
		callerID := c.GetString("user_id")
		caller, err := db.GetUserByID(c.Request.Context(), callerID)
		if err != nil || caller.Email != superAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the superadmin can change roles"})
			return
		}

		var req struct {
			Role string `json:"role" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if req.Role != "teacher" && req.Role != "student" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be 'teacher' or 'student'"})
			return
		}

		// Prevent demoting the superadmin
		target, err := db.GetUserByID(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if target.Email == superAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot change the superadmin's role"})
			return
		}

		if err := db.SetUserRole(c.Request.Context(), target.ID, models.Role(req.Role)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role updated to " + req.Role})
	}
}

func AdminListApps(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := db.ListAllApps(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch apps"})
			return
		}
		if apps == nil {
			apps = []*models.App{}
		}
		c.JSON(http.StatusOK, apps)
	}
}

func AdminHideApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Hidden bool `json:"hidden"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := db.SetAppHidden(c.Request.Context(), c.Param("id"), req.Hidden); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
			return
		}
		msg := "App is now visible"
		if req.Hidden {
			msg = "App hidden"
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}
