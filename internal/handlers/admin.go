package handlers

import (
	"net/http"

	"reploy/internal/models"
	"reploy/internal/storage"

	"github.com/gin-gonic/gin"
)

func AdminListUsers(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := db.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法取得使用者列表"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效請求"})
			return
		}
		if err := db.SetUserBanned(c.Request.Context(), c.Param("id"), req.Banned); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失敗"})
			return
		}
		msg := "已解封使用者"
		if req.Banned {
			msg = "已封禁使用者"
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}

func AdminListApps(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := db.ListAllApps(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法取得應用程式列表"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效請求"})
			return
		}
		if err := db.SetAppHidden(c.Request.Context(), c.Param("id"), req.Hidden); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失敗"})
			return
		}
		msg := "已顯示應用程式"
		if req.Hidden {
			msg = "已隱藏應用程式"
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}
