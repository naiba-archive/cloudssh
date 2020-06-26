package router

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
	"golang.org/x/crypto/bcrypt"

	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/validator"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// Serve ..
func Serve(conf string, port int) {
	if err := dao.InitConfig(conf); err != nil {
		panic(err)
	}
	if err := dao.InitDB(dao.Conf.DBDSN); err != nil {
		panic(err)
	}
	if dao.Conf.Debug {
		dao.DB = dao.DB.Debug()
	}
	if err := dao.DB.AutoMigrate(&model.User{}, &model.Organization{}, &model.Server{}, &model.OrganizationMember{}, &model.ServerOrganization{}).Error; err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Settings.ErrorHandler = func(c *fiber.Ctx, err error) {
		c.JSON(apiio.Response{
			Success: false,
			Message: err.Error(),
		})
	}

	app.Use(timer())
	app.Use(logger.New())

	app.Use(func(c *fiber.Ctx) {
		authHeader := c.Get("Authorization")
		arr := strings.Split(authHeader, " ")
		if len(arr) == 2 {
			var user model.User
			if dao.DB.First(&user, "token = ? AND token_expires > ?", arr[1], time.Now()).Error == nil {
				c.Locals("user", user)
			}
		}
		c.Next()
	})

	user := app.Group("/user", func(c *fiber.Ctx) {
		if c.Locals("user") == nil {
			c.Next(errors.New("You must login to continue"))
			return
		}
		c.Next()
	})

	user.Get("/logout", func(c *fiber.Ctx) {
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
	})

	user.Post("/server", func(c *fiber.Ctx) {
		user := c.Locals("user").(model.User)
		var req apiio.NewServerRequest
		if err := c.BodyParser(&req); err != nil {
			c.Next(err)
			return
		}
		if err := validator.Validator.Struct(req); err != nil {
			c.Next(err)
			return
		}

		var server model.Server
		if req.OrganizationID > 0 {
			var count uint64
			dao.DB.Where(&model.OrganizationMember{}, "user_id = ? AND organization_id = ?", user.ID, req.OrganizationID).Count(&count)
			if count == 0 {
				c.Next(fmt.Errorf("You don't have permission to write organization(%d)", req.OrganizationID))
				return
			}
			server.OwnerType = model.ServerOwnerTypeOrganization
			server.OwnerID = req.OrganizationID
		} else {
			server.OwnerType = model.ServerOwnerTypeUser
			server.OwnerID = user.ID
		}

		server.IP = req.IP
		server.Key = req.Key
		server.LoginWith = req.LoginWith
		server.Name = req.Name
		server.Port = req.Port
		server.User = req.User

		if err := dao.DB.Save(&server).Error; err != nil {
			c.Next(err)
			return
		}

		c.JSON(apiio.Response{
			Success: true,
			Message: fmt.Sprintf("Add server successful %s(%d)", req.Name, server.ID),
		})
	})

	user.Post("/server/batch-delete", func(c *fiber.Ctx) {
		user := c.Locals("user").(model.User)
		var req apiio.DeleteServerRequest
		if err := c.BodyParser(&req); err != nil {
			c.Next(err)
			return
		}
		if err := validator.Validator.Struct(req); err != nil {
			c.Next(err)
			return
		}

		var originCount int
		for i := 0; i < len(req.ID); i++ {
			if req.ID[i] != 0 {
				originCount++
			}
		}
		if originCount == 0 {
			c.Next(errors.New("empty server list"))
			return
		}

		var organizationID dao.FindIDResp
		dao.DB.Model(&model.OrganizationMember{}).Select("organization_id as id").Where("user_id = ?", user.ID).Scan(&organizationID)
		var dbCount int
		dao.DB.Model(&model.Server{}).Where("((owner_type = ? AND owner_id = ?) OR (owner_type = ? AND owner_id in (?))) AND id in (?)", model.ServerOwnerTypeUser, user.ID, model.ServerOwnerTypeOrganization, organizationID.ID, req.ID).Count(&dbCount)
		if dbCount != originCount {
			c.Next(errors.New("Some server not belongs you"))
			return
		}

		if err := dao.DB.Delete(&model.Server{}, "id in (?)", req.ID).Error; err != nil {
			c.Next(err)
			return
		}

		c.JSON(apiio.Response{
			Success: true,
			Message: fmt.Sprintf("Delete servers (%v) successful!", req.ID),
		})
	})

	user.Get("/server", func(c *fiber.Ctx) {
		user := c.Locals("user").(model.User)
		var organizationID dao.FindIDResp
		dao.DB.Model(&model.OrganizationMember{}).Select("organization_id as id").Where("user_id = ?", user.ID).Scan(&organizationID)
		var servers []model.Server
		dao.DB.Find(&servers, "(owner_type = ? AND owner_id = ?) OR (owner_type = ? AND owner_id in (?))", model.ServerOwnerTypeUser, user.ID, model.ServerOwnerTypeOrganization, organizationID.ID)
		c.JSON(apiio.ListServerResponse{
			Response: apiio.Response{
				Success: true,
				Message: "",
			},
			Data: servers,
		})
	})

	app.Post("/signup", func(c *fiber.Ctx) {
		var req apiio.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			c.Next(err)
			return
		}
		if err := validator.Validator.Struct(req); err != nil {
			c.Next(err)
			return
		}

		if _, err := xcrypto.BytesToPublicKey([]byte(req.Pubkey)); err != nil {
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
	})

	app.Post("/login", func(c *fiber.Ctx) {
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
	})

	app.Listen(port)
}

func timer() func(*fiber.Ctx) {
	return func(c *fiber.Ctx) {
		// start timer
		start := time.Now()
		// next routes
		c.Next()
		// stop timer
		stop := time.Now()
		// Do something with response
		c.Append("Server-Timing", fmt.Sprintf("app;dur=%v", stop.Sub(start).String()))
	}
}
