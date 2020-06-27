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

// GetServer ..
func GetServer(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var organizationID []dao.FindIDResp
	dao.DB.Model(&model.OrganizationUser{}).Select("organization_id as id").Where("user_id = ?", user.ID).Scan(&organizationID)
	var ids []int64
	for i := 0; i < len(organizationID); i++ {
		ids = append(ids, organizationID[i].ID)
	}
	var server model.Server
	if err := dao.DB.First(&server, "((owner_type = ? AND owner_id = ?) OR (owner_type = ? AND owner_id in (?))) AND id = ?", model.ServerOwnerTypeUser, user.ID, model.ServerOwnerTypeOrganization, ids, c.Params("id")).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.GetServerResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: server,
	})
}

// EditServer ..
func EditServer(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var req apiio.ServerRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	var orgUser model.OrganizationUser
	if req.OrganizationID > 0 {
		if err := dao.DB.First(&orgUser, "organization_id = ? AND user_id = ? AND permission >= ?", req.OrganizationID, user.ID, model.OUPermissionReadWrite).Error; err != nil {
			c.Next(err)
			return
		}
	}
	var server model.Server
	if err := dao.DB.First(&server, "((owner_type = ? AND owner_id = ?) OR (owner_type = ? AND owner_id = ?)) AND id = ?", model.ServerOwnerTypeUser, user.ID, model.ServerOwnerTypeOrganization, orgUser.OrganizationID, c.Params("id")).Error; err != nil {
		c.Next(err)
		return
	}

	server.IP = req.IP
	server.Key = req.Key
	server.LoginWith = req.LoginWith
	server.Name = req.Name
	server.Port = req.Port
	server.LoginUser = req.LoginUser

	if err := dao.DB.Save(&server).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: fmt.Sprintf("Edit server successful %s(%d)", req.Name, server.ID),
	})
}

// ListServer ..
func ListServer(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var servers []model.Server
	dao.DB.Find(&servers, "owner_type = ? AND owner_id = ?", model.ServerOwnerTypeUser, user.ID)
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

	var dbCount int
	if req.OrganizationID > 0 {
		var orgUser model.OrganizationUser
		if err := dao.DB.First(&orgUser, "permission >= ? AND organization_id = ? AND user_id = ?", model.OUPermissionReadWrite, req.OrganizationID, user.ID).Error; err != nil {
			c.Next(err)
			return
		}
		dao.DB.Model(&model.Server{}).Where("owner_type = ? AND owner_id = ? AND id in (?)", model.ServerOwnerTypeOrganization, req.OrganizationID, req.ID).Count(&dbCount)
	} else {
		dao.DB.Model(&model.Server{}).Where("owner_type = ? AND owner_id = ? AND id in (?)", model.ServerOwnerTypeUser, user.ID, req.ID).Count(&dbCount)
	}
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
	var req apiio.ServerRequest
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
		dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ? AND permission >= ?", user.ID, req.OrganizationID, model.OUPermissionReadWrite).Count(&count)
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
	server.LoginUser = req.LoginUser

	if err := dao.DB.Save(&server).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: fmt.Sprintf("Add server successful %s(%d)", req.Name, server.ID),
	})
}
