package middleware

import (
	"errors"
	"strings"

	"eka-dev.cloud/sse-gateway/config"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type PermissionAction struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

type Claims struct {
	FullName    string                      `json:"FullName"`
	Email       string                      `json:"Email"`
	UserId      int64                       `json:"UserId"`
	Type        string                      `json:"Type"`
	Role        string                      `json:"Role"`
	RoleId      int                         `json:"RoleId"`
	Permissions map[string]PermissionAction `json:"Permissions"`
	jwt.RegisteredClaims
}

func getJwtKey() []byte {
	if config.Config.SecretJwt != "" {
		return []byte(config.Config.SecretJwt)
	}
	return []byte("super-secret-jwt-key")
}

func getTokenFromHeader(c *fiber.Ctx) string {
	bearer := c.Get("Authorization")
	if bearer == "" {
		return ""
	}
	if strings.HasPrefix(bearer, "Bearer ") {
		return bearer[len("Bearer "):]
	}
	return bearer
}

func validateToken(c *fiber.Ctx) (*Claims, error) {
	tokenString := getTokenFromHeader(c)
	if tokenString == "" {
		return nil, response.Unauthorized("Missing Token", nil)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, response.Unauthorized("Unexpected signing method", nil)
		}
		return getJwtKey(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, response.Unauthorized("Token expired", nil)
		}
		return nil, response.Unauthorized("Invalid token", nil)
	}

	if !token.Valid {
		return nil, response.Unauthorized("Invalid token", nil)
	}

	if strings.ToLower(claims.Type) != "access" {
		return nil, response.Unauthorized("Invalid token type", nil)
	}

	return claims, nil
}

func ValidateTokenQuery(c *fiber.Ctx) (*Claims, error) {
	tokenString := c.Query("token")
	if tokenString == "" {
		tokenString = c.Cookies("token")
	}
	if tokenString == "" {
		tokenString = getTokenFromHeader(c)
	}
	if tokenString == "" {
		return nil, response.Unauthorized("Missing Token", nil)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, response.Unauthorized("Unexpected signing method", nil)
		}
		return getJwtKey(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, response.Unauthorized("Token expired", nil)
		}
		return nil, response.Unauthorized("Invalid token", nil)
	}

	if !token.Valid {
		return nil, response.Unauthorized("Invalid token", nil)
	}

	if strings.ToLower(claims.Type) != "access" {
		return nil, response.Unauthorized("Invalid token type", nil)
	}

	return claims, nil
}

func RequireAuth(c *fiber.Ctx) error {
	claims, err := validateToken(c)
	if err != nil {
		var appErr *response.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return response.Unauthorized("Unauthorized", nil)
	}
	c.Locals("user", claims)
	return c.Next()
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := validateToken(c)
		if err != nil {
			var appErr *response.AppError
			if errors.As(err, &appErr) {
				return err
			}
			return response.Unauthorized("Unauthorized", nil)
		}

		userRole := claims.Role
		for _, role := range roles {
			if strings.EqualFold(userRole, role) {
				c.Locals("user", claims)
				return c.Next()
			}
		}
		return response.Forbidden("Forbidden", nil)
	}
}
