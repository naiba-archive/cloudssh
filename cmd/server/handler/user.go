package handler

import (
	"github.com/gofiber/fiber"
	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
)

// GetUserInfo ..
func GetUserInfo(c *fiber.Ctx) {
	var user model.User
	dao.DB.First(&user, "id = ?", c.Params("id"))
	c.JSON(apiio.UserInfoResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct{ Pubkey string }{
			Pubkey: user.Pubkey,
		},
	})
}
