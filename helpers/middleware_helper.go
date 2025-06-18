package helpers

import (
    "fmt"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "Authorization header is required",
            })
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "Invalid authorization format. Use: Bearer <token>",
            })
        }

        tokenString := bearerToken[1]

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            
            return GetPublicKey(), nil
        })

        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "Invalid token",
                "error": err.Error(),
            })
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            if exp, ok := claims["exp"].(float64); ok {
                if jwt.NewNumericDate(time.Now()).Unix() > int64(exp) {
                    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                        "success": false,
                        "message": "Token has expired",
                    })
                }
            }

            c.Locals("claims", claims)
            
            if username, exists := claims["username"]; exists {
                c.Locals("username", username)
            }
            
            return c.Next()
        }

        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Invalid token claims",
        })
    }
}