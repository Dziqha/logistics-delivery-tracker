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

        // Format: Bearer <token>
        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "Invalid authorization format. Use: Bearer <token>",
            })
        }

        tokenString := bearerToken[1]

        // Parse JWT token with ECDSA public key
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Validate the signing method is ECDSA
            if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            
            // Return the ECDSA public key for verification
            return GetPublicKey(), nil
        })

        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "Invalid token",
                "error": err.Error(),
            })
        }

        // Validate token and extract claims
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            // Check token expiration
            if exp, ok := claims["exp"].(float64); ok {
                if jwt.NewNumericDate(time.Now()).Unix() > int64(exp) {
                    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                        "success": false,
                        "message": "Token has expired",
                    })
                }
            }

            // Store claims in context for use in handlers
            c.Locals("claims", claims)
            
            // Extract and store username for easier access
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