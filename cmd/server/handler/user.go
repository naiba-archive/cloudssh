package handler

import (
	"errors"

	"github.com/gofiber/fiber"
	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

// GetUserInfo ..
func GetUserInfo(c *fiber.Ctx) {
	var user model.User
	dao.DB.First(&user, "email = ?", c.Query("email"))
	c.JSON(apiio.UserInfoResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct{ Pubkey string }{
			Pubkey: user.Pubkey,
		},
	})
}

// ListAllTeamUser ..
func ListAllTeamUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var userTeam []model.TeamUser
	if err := dao.DB.Where("user_id = ?", user.ID).Find(&userTeam).Error; err != nil {
		c.Next(err)
		return
	}
	c.JSON(apiio.ListTeamUserResponse{
		Response: apiio.Response{
			Success: true,
			Message: "",
		},
		Data: struct {
			User  []model.TeamUser
			Key   map[uint64]string
			Email map[uint64]string
		}{
			User: userTeam,
		},
	})
}

// Passwd ..
func Passwd(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.PasswdRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPasswordHash))
	if err != nil {
		c.Next(err)
		return
	}

	ph, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), 14)
	if err != nil {
		c.Next(err)
		return
	}
	user.PasswordHash = string(ph)
	user.EncryptKey = req.EncryptKey
	user.Privatekey = req.Privatekey
	user.Pubkey = req.Pubkey

	tx := dao.DB.Begin()

	var count int
	tx.Model(&model.TeamUser{}).Where("user_id = ?", user.ID).Count(&count)
	if count != len(req.TeamUser) {
		c.Next(errors.New("Sync TeamUser data count not match"))
		return
	}
	tx.Model(&model.Server{}).Where("owner_id = ? AND owner_type = ?", user.ID, model.ServerOwnerTypeUser).Count(&count)
	if count != len(req.Server) {
		c.Next(errors.New("Sync Server data count not match"))
		return
	}
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}
	for i := 0; i < len(req.TeamUser); i++ {
		if err := tx.Model(&model.TeamUser{}).Where("team_id = ? AND user_id = ?", req.TeamUser[i].TeamID, user.ID).Update("private_key", req.TeamUser[i].PrivateKey).Error; err != nil {
			tx.Rollback()
			c.Next(err)
			return
		}
	}
	for i := 0; i < len(req.Server); i++ {
		req.Server[i].OwnerID = user.ID
		req.Server[i].OwnerType = model.ServerOwnerTypeUser
		if err := tx.Save(&req.Server[i]).Error; err != nil {
			tx.Rollback()
			c.Next(err)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	c.JSON(apiio.UserResponse{
		Response: apiio.Response{
			Success: true,
			Message: "Change password and sync data complated",
		},
		Data: user,
	})
}
