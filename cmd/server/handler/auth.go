package handler

import (
	"time"

	"github.com/gofiber/fiber"
	"golang.org/x/crypto/bcrypt"

	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/validator"
)

// Logout ..
func Logout(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	if err := user.RefreshToken(); err != nil {
		c.Next(err)
		return
	}
	user.TokenExpires = time.Unix(0, 0)
	if err := dao.DB.Save(&user).Error; err != nil {
		c.Next(err)
		return
	}
	c.JSON(apiio.Response{
		Success: true,
		Message: "logout successful!",
	})
}

// SignUp ..
func SignUp(c *fiber.Ctx) {
	var req apiio.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	var user model.User
	user.Email = req.Email
	var ph []byte
	ph, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), 14)
	user.PasswordHash = string(ph)
	user.EncryptKey = req.EncryptKey
	user.Privatekey = req.Privatekey
	user.Pubkey = req.Pubkey
	if err != nil {
		c.Next(err)
		return
	}
	if err := user.RefreshToken(); err != nil {
		c.Next(err)
		return
	}
	if err := dao.DB.Save(&user).Error; err != nil {
		c.Next(err)
		return
	}
	c.JSON(apiio.RegisterResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: user,
	})
}

// Login ..
func Login(c *fiber.Ctx) {
	var req apiio.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}
	var user model.User
	if err := dao.DB.First(&user, "email = ?", req.Email).Error; err != nil {
		c.Next(err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.PasswordHash)); err != nil {
		c.Next(err)
		return
	}
	if err := user.RefreshToken(); err != nil {
		c.Next(err)
		return
	}
	if err := dao.DB.Save(&user).Error; err != nil {
		c.Next(err)
		return
	}
	c.JSON(apiio.RegisterResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: user,
	})
}
