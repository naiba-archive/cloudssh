package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/validator"
)

// ListServer ..
func ListServer(c *fiber.Ctx) {
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
}

// BatchDelete ..
func BatchDelete(c *fiber.Ctx) {
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
}

// CreateServer ..
func CreateServer(c *fiber.Ctx) {
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
}
