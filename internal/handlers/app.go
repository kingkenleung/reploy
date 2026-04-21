package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"reploy/internal/models"
	"reploy/internal/storage"

	"github.com/gin-gonic/gin"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

const maxHTMLSize = 500 * 1024 // 500KB

type createAppRequest struct {
	Slug        string          `json:"slug" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	HTMLContent string          `json:"html_content" binding:"required"`
	Category    json.RawMessage `json:"category"`
}

type updateAppRequest struct {
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	HTMLContent string          `json:"html_content" binding:"required"`
	Category    json.RawMessage `json:"category"`
}

func ListApps(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		apps, err := db.ListAppsByUser(c.Request.Context(), userID)
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

func GetApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		app, err := db.GetAppByID(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到應用程式"})
			return
		}
		if app.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "無權限"})
			return
		}
		c.JSON(http.StatusOK, app)
	}
}

func CreateApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createAppRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "請填寫所有必填欄位"})
			return
		}

		req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		if !slugRegex.MatchString(req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "連結名稱只能包含小寫英文字母、數字和連字號"})
			return
		}
		if len(req.HTMLContent) > maxHTMLSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "HTML 檔案不能超過 500KB"})
			return
		}

		app, err := db.CreateApp(c.Request.Context(), &models.App{
			UserID:      c.GetString("user_id"),
			Slug:        req.Slug,
			Title:       req.Title,
			Description: req.Description,
			HTMLContent: req.HTMLContent,
			Category:    req.Category,
		})
		if err != nil {
			if strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusConflict, gin.H{"error": "此連結名稱已被使用，請換一個"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "建立失敗"})
			return
		}
		c.JSON(http.StatusCreated, app)
	}
}

func UpdateApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		role := c.GetString("role")

		app, err := db.GetAppByID(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到應用程式"})
			return
		}
		// Only owner or teacher can update
		if app.UserID != userID && role != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"error": "無權限"})
			return
		}

		var req updateAppRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "請填寫所有必填欄位"})
			return
		}
		if len(req.HTMLContent) > maxHTMLSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "HTML 檔案不能超過 500KB"})
			return
		}

		if err := db.UpdateAppContent(c.Request.Context(), app.ID, req.Title, req.Description, req.HTMLContent, req.Category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存失敗"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已儲存"})
	}
}

func DeleteApp(db *storage.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		app, err := db.GetAppByID(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到應用程式"})
			return
		}
		if app.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "無權限"})
			return
		}
		if err := db.DeleteApp(c.Request.Context(), app.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已刪除"})
	}
}
