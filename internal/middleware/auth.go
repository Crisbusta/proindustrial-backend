package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Routes exempt from the mustChangePassword check.
// Las rutas de admin también deben estar aquí: sin ellas, un admin con el
// flag activo recibía 403 en todo, incluida la única ruta que le permitiría
// desbloquearse — justo el mecanismo necesario para forzar una rotación.
var passwordChangeExemptRoutes = map[string]bool{
	"/api/auth/change-password":       true,
	"/api/auth/me":                    true,
	"/api/admin/auth/change-password": true,
	"/api/admin/auth/me":              true,
}

const UserIDKey = "userID"
const CompanyIDKey = "companyID"
const RoleKey = "role"

func Auth(jwtSecret string, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid role"})
			return
		}

		if len(allowedRoles) > 0 {
			allowed := false
			for _, candidate := range allowedRoles {
				if role == candidate {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}

		c.Set(UserIDKey, userID)
		c.Set(RoleKey, role)
		if companyID, ok := claims["companyId"].(string); ok && companyID != "" {
			c.Set(CompanyIDKey, companyID)
		}

		if mustChange, ok := claims["mustChangePassword"].(bool); ok && mustChange {
			if !passwordChangeExemptRoutes[c.FullPath()] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "debe cambiar su contraseña antes de continuar"})
				return
			}
		}

		c.Next()
	}
}
