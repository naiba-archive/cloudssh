package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber"

	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/model"
)

// Auth ..
func Auth(c *fiber.Ctx) {
	authHeader := c.Get("Authorization")
	arr := strings.Split(authHeader, " ")
	if len(arr) == 2 {
		var user model.User
		if dao.DB.First(&user, "token = ? AND token_expires > ?", arr[1], time.Now()).Error == nil {
			c.Locals("user", user)
		}
	}
	c.Next()
}

// Protected ..
func Protected(c *fiber.Ctx) {
	if c.Locals("user") == nil {
		c.Next(errors.New("You must login to continue"))
		return
	}
	c.Next()
}
