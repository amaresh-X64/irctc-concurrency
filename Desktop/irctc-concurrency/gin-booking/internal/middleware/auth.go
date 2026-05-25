package middleware

import (
	"net/http"
	"strings"

	"gin-booking/internal/constants"
	"gin-booking/internal/helpers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			helpers.ErrorResponse(c, http.StatusUnauthorized, constants.MsgUnauthorized, nil)
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				// reject anything that isn't HMAC (blocks alg:none, RS256, etc.)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(constants.JWTSecretKey), nil
			},
			jwt.WithValidMethods([]string{constants.JWTAlgorithm}),
		)

		if err != nil || !token.Valid {
			helpers.ErrorResponse(c, http.StatusUnauthorized, constants.MsgTokenExpired, nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			helpers.ErrorResponse(c, http.StatusUnauthorized, constants.MsgUnauthorized, nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims["sub"])
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}