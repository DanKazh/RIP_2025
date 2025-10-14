package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"rip2025/internal/app/ds"
	"rip2025/internal/app/redis"
	"rip2025/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const cookieName = "harvest_jwt"

func WithAuthCheck(jwtSecret string, redisClient *redis.Client, assignedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("=== COOKIES: %+v\n", ctx.Request.Cookies())
		allowsGuest := false
		for _, r := range assignedRoles {
			if r == role.Guest {
				allowsGuest = true
				break
			}
		}

		cookie, err := ctx.Request.Cookie(cookieName)
		if err != nil {
			if allowsGuest {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		tokenStr := strings.TrimSpace(cookie.Value)
		if tokenStr == "" {
			if allowsGuest {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		err = redisClient.CheckJWTInBlacklist(ctx, tokenStr)
		if err == nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token blacklisted"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			if allowsGuest {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok || !token.Valid {
			if allowsGuest {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if claims.ExpiresAt < time.Now().Unix() {
			if allowsGuest {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		userRole := role.FromString(claims.Role)
		allowed := false
		for _, r := range assignedRoles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized",
				"role":  claims.Role,
			})
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("user_role", claims.Role)

		ctx.Next()
	}
}
