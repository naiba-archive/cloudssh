package handler

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/validator"
)

// ListOrganizationServer ..
func ListOrganizationServer(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var userOrg model.OrganizationUser
	if err := dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ?", user.ID, c.Params("id")).First(&userOrg).Error; err != nil {
		c.Next(err)
		return
	}
	var servers []model.Server
	dao.DB.Find(&servers, "owner_type = ? AND owner_id = ?", model.ServerOwnerTypeOrganization, c.Params("id"))
	c.JSON(apiio.ListServerResponse{
		Response: apiio.Response{
			Success: true,
			Message: "",
		},
		Data: servers,
	})
}

// AddOrganizationUser ..
func AddOrganizationUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.AddOrganizationUserRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	var count uint64
	dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ? AND permission >= ?", user.ID, req.OrganizationID, model.OUPermissionOwner).Count(&count)
	if count == 0 {
		c.Next(fmt.Errorf("You don't have permission to manage organization(%d)", req.OrganizationID))
		return
	}
}

// GetOrganization ..
func GetOrganization(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var orgUser model.OrganizationUser
	if err := dao.DB.First(&orgUser, "organization_id = ? AND user_id = ?", c.Params("id"), user.ID).Error; err != nil {
		c.Next(err)
		return
	}

	var org model.Organization
	if err := dao.DB.First(&org, "id = ?", orgUser.OrganizationID).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.GetOrganizationResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct {
			Organization     model.Organization
			OrganizationUser model.OrganizationUser
		}{
			Organization:     org,
			OrganizationUser: orgUser,
		},
	})
}

// CreateOrg ..
func CreateOrg(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.OrgRequrest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	tx := dao.DB.Begin()

	var org model.Organization
	org.Name = req.Name
	org.Pubkey = req.Pubkey
	if err := tx.Save(&org).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	var om model.OrganizationUser
	om.OrganizationID = org.ID
	om.Permission = model.OUPermissionReadWrite
	om.PrivateKey = req.Prikey
	om.UserID = user.ID
	if err := tx.Save(&om).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: fmt.Sprintf("Add server successful %s(%d)", req.Name, org.ID),
	})
}
