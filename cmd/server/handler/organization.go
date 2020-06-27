package handler

import (
	"errors"
	"fmt"
	"strconv"

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

// ListOrganizationUser ..
func ListOrganizationUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var userOrg model.OrganizationUser
	if err := dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ? AND permission >= ?", user.ID, c.Params("id"), model.OUPermissionOwner).First(&userOrg).Error; err != nil {
		c.Next(err)
		return
	}
	var users []model.OrganizationUser
	dao.DB.Find(&users, "organization_id = ?", c.Params("id"))
	var userIDs []uint64
	for i := 0; i < len(users); i++ {
		userIDs = append(userIDs, users[i].UserID)
	}
	email := make(map[uint64]string)
	var userPubkeys []model.User
	dao.DB.Select("id,email,pubkey").Find(&userPubkeys, "id in (?)", userIDs)
	keyData := make(map[uint64]string)
	for i := 0; i < len(userPubkeys); i++ {
		keyData[userPubkeys[i].ID] = userPubkeys[i].Pubkey
		email[userPubkeys[i].ID] = userPubkeys[i].Email
	}
	c.JSON(apiio.ListOrganizationUserResponse{
		Response: apiio.Response{
			Success: true,
			Message: "",
		},
		Data: struct {
			User  []model.OrganizationUser
			Key   map[uint64]string
			Email map[uint64]string
		}{
			User:  users,
			Key:   keyData,
			Email: email,
		},
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

	orgID, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if orgID == 0 {
		c.Next(errors.New("invalid organization id"))
		return
	}

	var count uint64
	dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ? AND permission >= ?", user.ID, orgID, model.OUPermissionOwner).Count(&count)
	if count == 0 {
		c.Next(fmt.Errorf("You don't have permission to manage organization(%d)", orgID))
		return
	}

	var u model.User
	if err := dao.DB.Select("id").First(&u, "email = ?", req.Email).Error; err != nil {
		c.Next(err)
		return
	}

	var ou model.OrganizationUser
	ou.OrganizationID = orgID
	ou.UserID = u.ID
	ou.Permission = req.Permission
	ou.PrivateKey = req.Prikey

	if err := dao.DB.Save(&ou).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: "Add user to organization successful",
	})
}

// BatchDeleteOrganizationUser ..
func BatchDeleteOrganizationUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.DeleteOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	if err := dao.DB.First(&model.OrganizationUser{}, "user_id = ? AND organization_id = ? AND permission >= ?", user.ID, c.Params("id"), model.OUPermissionOwner).Error; err != nil {
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
		c.Next(errors.New("empty organization list"))
		return
	}

	var dbCount int
	dao.DB.Model(&model.OrganizationUser{}).Where("organization_id = ? AND user_id in (?) AND user_id != ?", c.Params("id"), req.ID, user.ID).Count(&dbCount)

	if dbCount != originCount {
		c.Next(errors.New("Some organization not belongs you"))
		return
	}

	if err := dao.DB.Delete(&model.OrganizationUser{}, "organization_id = ? AND user_id in (?)", c.Params("id"), req.ID).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: fmt.Sprintf("Delete organizations (%v) successful!", req.ID),
	})
}

// BatchDeleteOrganization ..
func BatchDeleteOrganization(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.DeleteOrganizationRequest
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
		c.Next(errors.New("empty organization list"))
		return
	}

	var dbCount int
	dao.DB.Model(&model.OrganizationUser{}).Where("permission >= ? AND organization_id in (?) AND user_id = ?", model.OUPermissionOwner, req.ID, user.ID).Count(&dbCount)

	if dbCount != originCount {
		c.Next(errors.New("Some organization not belongs you"))
		return
	}

	tx := dao.DB.Begin()

	if err := tx.Delete(&model.Organization{}, "id in (?)", req.ID).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	if err := tx.Delete(&model.Server{}, "owner_type = ? AND owner_id in (?)", model.ServerOwnerTypeOrganization, req.ID).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	if err := tx.Delete(&model.OrganizationUser{}, "organization_id in (?)", req.ID).Error; err != nil {
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
		Message: fmt.Sprintf("Delete organizations (%v) successful!", req.ID),
	})
}

// UpdateOrganization ..
func UpdateOrganization(c *fiber.Ctx) {
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

	organizationID, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if organizationID == 0 {
		c.Next(errors.New("invaild organization id"))
		return
	}

	var count uint64
	dao.DB.Model(&model.OrganizationUser{}).Where("user_id = ? AND organization_id = ? AND permission >= ?", user.ID, organizationID, model.OUPermissionOwner).Count(&count)
	if count == 0 {
		c.Next(fmt.Errorf("You don't have permission to manage organization(%d)", organizationID))
		return
	}
	var org model.Organization
	if err := dao.DB.First(&org, "id = ?", organizationID).Error; err != nil {
		c.Next(err)
		return
	}

	tx := dao.DB.Begin()
	if req.Pubkey != "" && org.Pubkey != req.Pubkey {
		// reset all organization data
		var userIDs []uint64
		for i := 0; i < len(req.Users); i++ {
			userIDs = append(userIDs, req.Users[i].UserID)
			if req.Users[i].OrganizationID != organizationID {
				c.Next(errors.New("user organization id missmatch"))
				return
			}
		}
		var count int
		tx.Model(&model.OrganizationUser{}).Where("organization_id = ? AND user_id in (?)", organizationID, userIDs).Count(&count)
		if count != len(userIDs) {
			c.Next(errors.New("user num missmatch"))
			return
		}
		var serverIDs []uint64
		for i := 0; i < len(req.Servers); i++ {
			serverIDs = append(serverIDs, req.Servers[i].ID)
			if req.Servers[i].OwnerType != model.ServerOwnerTypeOrganization || req.Servers[i].OwnerID != organizationID {
				c.Next(errors.New("server organization id missmatch"))
				return
			}
		}
		tx.Model(&model.Server{}).Where("owner_type = ? AND owner_id = ? AND id in (?)", model.ServerOwnerTypeOrganization, c.Params("id"), serverIDs).Count(&count)
		if count != len(serverIDs) {
			c.Next(errors.New("server num missmatch"))
			return
		}
		for i := 0; i < len(req.Users); i++ {
			if err := tx.Save(&req.Users[i]).Error; err != nil {
				tx.Rollback()
				c.Next(err)
				return
			}
		}
		for i := 0; i < len(req.Servers); i++ {
			if err := tx.Save(&req.Servers[i]).Error; err != nil {
				tx.Rollback()
				c.Next(err)
				return
			}
		}
		org.Pubkey = req.Pubkey
	}
	org.Name = req.Name
	if err := tx.Save(&org).Error; err != nil {
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
		Message: "Organization update successful!",
	})
}

// ListOrganization ..
func ListOrganization(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var ous []model.OrganizationUser
	dao.DB.Find(&ous, "user_id = ?", user.ID)
	permission := make(map[uint64]uint64)
	var ids []uint64
	for i := 0; i < len(ous); i++ {
		ids = append(ids, ous[i].OrganizationID)
		permission[ous[i].OrganizationID] = ous[i].Permission
	}
	var orgs []model.Organization
	dao.DB.Find(&orgs, "id in (?)", ids)

	c.JSON(apiio.ListOrganizationResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct {
			Orgnazation []model.Organization
			Permission  map[uint64]uint64
		}{
			Orgnazation: orgs,
			Permission:  permission,
		},
	})
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
	var req apiio.NewOrgRequrest
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
	om.Permission = model.OUPermissionOwner
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
