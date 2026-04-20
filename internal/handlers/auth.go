package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"reploy/internal/auth"
	"reploy/internal/middleware"
	"reploy/internal/models"
	"reploy/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var oauthStateStore = map[string]bool{}

func GoogleLogin(oauthCfg *auth.OAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := randomState()
		oauthStateStore[state] = true
		c.Redirect(http.StatusTemporaryRedirect, oauthCfg.AuthCodeURL(state))
	}
}

func GoogleCallback(db *storage.DB, oauthCfg *auth.OAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		if !oauthStateStore[state] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 OAuth state"})
			return
		}
		delete(oauthStateStore, state)

		userInfo, err := auth.GetUserInfo(c.Request.Context(), oauthCfg.Config, c.Query("code"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法取得使用者資訊"})
			return
		}

		allowedDomain := os.Getenv("ALLOWED_EMAIL_DOMAIN")
		if !strings.HasSuffix(userInfo.Email, "@"+allowedDomain) {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"message": "只允許 @" + allowedDomain + " 的帳號登入",
			})
			return
		}

		pyccode := strings.TrimSuffix(userInfo.Email, "@"+allowedDomain)

		saved, err := db.UpsertUser(c.Request.Context(), &models.User{
			GoogleID:  userInfo.ID,
			Email:     userInfo.Email,
			PYCCode:   pyccode,
			Name:      userInfo.Name,
			AvatarURL: userInfo.Picture,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫錯誤"})
			return
		}

		tokenStr, err := issueJWT(saved)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法產生登入憑證"})
			return
		}

		c.SetCookie("token", tokenStr, 60*60*24*7, "/", "", false, true)
		c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetCookie("token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

func issueJWT(u *models.User) (string, error) {
	claims := middleware.Claims{
		UserID:   u.ID,
		PYCCode:  u.PYCCode,
		Role:     string(u.Role),
		IsBanned: u.IsBanned,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
