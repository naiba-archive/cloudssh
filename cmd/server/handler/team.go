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

// ListTeamServer ..
func ListTeamServer(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var userTeam model.TeamUser
	if err := dao.DB.Model(&model.TeamUser{}).Where("user_id = ? AND team_id = ?", user.ID, c.Params("id")).First(&userTeam).Error; err != nil {
		c.Next(err)
		return
	}
	var servers []model.Server
	dao.DB.Find(&servers, "owner_type = ? AND owner_id = ?", model.ServerOwnerTypeTeam, c.Params("id"))
	c.JSON(apiio.ListServerResponse{
		Response: apiio.Response{
			Success: true,
			Message: "",
		},
		Data: servers,
	})
}

// ListTeamUser ..
func ListTeamUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var userTeam model.TeamUser
	if err := dao.DB.Model(&model.TeamUser{}).Where("user_id = ? AND team_id = ? AND permission >= ?", user.ID, c.Params("id"), model.OUPermissionOwner).First(&userTeam).Error; err != nil {
		c.Next(err)
		return
	}
	var users []model.TeamUser
	dao.DB.Find(&users, "team_id = ?", c.Params("id"))
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
			User:  users,
			Key:   keyData,
			Email: email,
		},
	})
}

// AddTeamUser ..
func AddTeamUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.AddTeamUserRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	teamID, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if teamID == 0 {
		c.Next(errors.New("invalid team id"))
		return
	}

	var count uint64
	dao.DB.Model(&model.TeamUser{}).Where("user_id = ? AND team_id = ? AND permission >= ?", user.ID, teamID, model.OUPermissionOwner).Count(&count)
	if count == 0 {
		c.Next(fmt.Errorf("You don't have permission to manage team(%d)", teamID))
		return
	}

	var u model.User
	if err := dao.DB.Select("id").First(&u, "email = ?", req.Email).Error; err != nil {
		c.Next(err)
		return
	}

	var ou model.TeamUser
	ou.TeamID = teamID
	ou.UserID = u.ID
	ou.Permission = req.Permission
	ou.PrivateKey = req.Prikey

	if err := dao.DB.Save(&ou).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: "Add user to team successful",
	})
}

// BatchDeleteTeamUser ..
func BatchDeleteTeamUser(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.DeleteTeamRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	if err := dao.DB.First(&model.TeamUser{}, "user_id = ? AND team_id = ? AND permission >= ?", user.ID, c.Params("id"), model.OUPermissionOwner).Error; err != nil {
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
		c.Next(errors.New("empty team list"))
		return
	}

	var dbCount int
	dao.DB.Model(&model.TeamUser{}).Where("team_id = ? AND user_id in (?) AND user_id != ?", c.Params("id"), req.ID, user.ID).Count(&dbCount)

	if dbCount != originCount {
		c.Next(errors.New("Some team not belongs you"))
		return
	}

	if err := dao.DB.Delete(&model.TeamUser{}, "team_id = ? AND user_id in (?)", c.Params("id"), req.ID).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.Response{
		Success: true,
		Message: fmt.Sprintf("Delete teams (%v) successful!", req.ID),
	})
}

// BatchDeleteTeam ..
func BatchDeleteTeam(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.DeleteTeamRequest
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
		c.Next(errors.New("empty team list"))
		return
	}

	var dbCount int
	dao.DB.Model(&model.TeamUser{}).Where("permission >= ? AND team_id in (?) AND user_id = ?", model.OUPermissionOwner, req.ID, user.ID).Count(&dbCount)

	if dbCount != originCount {
		c.Next(errors.New("Some team not belongs you"))
		return
	}

	tx := dao.DB.Begin()

	if err := tx.Delete(&model.Team{}, "id in (?)", req.ID).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	if err := tx.Delete(&model.Server{}, "owner_type = ? AND owner_id in (?)", model.ServerOwnerTypeTeam, req.ID).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	if err := tx.Delete(&model.TeamUser{}, "team_id in (?)", req.ID).Error; err != nil {
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
		Message: fmt.Sprintf("Delete teams (%v) successful!", req.ID),
	})
}

// UpdateTeam ..
func UpdateTeam(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.TeamRequrest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	teamID, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if teamID == 0 {
		c.Next(errors.New("invaild team id"))
		return
	}

	var count uint64
	dao.DB.Model(&model.TeamUser{}).Where("user_id = ? AND team_id = ? AND permission >= ?", user.ID, teamID, model.OUPermissionOwner).Count(&count)
	if count == 0 {
		c.Next(fmt.Errorf("You don't have permission to manage team(%d)", teamID))
		return
	}
	var team model.Team
	if err := dao.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.Next(err)
		return
	}

	tx := dao.DB.Begin()
	if req.Pubkey != "" && team.Pubkey != req.Pubkey {
		// reset all team data
		var userIDs []uint64
		for i := 0; i < len(req.Users); i++ {
			userIDs = append(userIDs, req.Users[i].UserID)
			if req.Users[i].TeamID != teamID {
				c.Next(errors.New("user team id missmatch"))
				return
			}
		}
		var count int
		tx.Model(&model.TeamUser{}).Where("team_id = ? AND user_id in (?)", teamID, userIDs).Count(&count)
		if count != len(userIDs) {
			c.Next(errors.New("user num missmatch"))
			return
		}
		var serverIDs []uint64
		for i := 0; i < len(req.Servers); i++ {
			serverIDs = append(serverIDs, req.Servers[i].ID)
			if req.Servers[i].OwnerType != model.ServerOwnerTypeTeam || req.Servers[i].OwnerID != teamID {
				c.Next(errors.New("server team id missmatch"))
				return
			}
		}
		tx.Model(&model.Server{}).Where("owner_type = ? AND owner_id = ? AND id in (?)", model.ServerOwnerTypeTeam, c.Params("id"), serverIDs).Count(&count)
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
		team.Pubkey = req.Pubkey
	}
	team.Name = req.Name
	if err := tx.Save(&team).Error; err != nil {
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
		Message: "Team update successful!",
	})
}

// ListTeam ..
func ListTeam(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var ous []model.TeamUser
	dao.DB.Find(&ous, "user_id = ?", user.ID)
	permission := make(map[uint64]uint64)
	var ids []uint64
	for i := 0; i < len(ous); i++ {
		ids = append(ids, ous[i].TeamID)
		permission[ous[i].TeamID] = ous[i].Permission
	}
	var teams []model.Team
	dao.DB.Find(&teams, "id in (?)", ids)

	c.JSON(apiio.ListTeamResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct {
			Teamnazation []model.Team
			Permission   map[uint64]uint64
		}{
			Teamnazation: teams,
			Permission:   permission,
		},
	})
}

// GetTeam ..
func GetTeam(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)

	var teamUser model.TeamUser
	if err := dao.DB.First(&teamUser, "team_id = ? AND user_id = ?", c.Params("id"), user.ID).Error; err != nil {
		c.Next(err)
		return
	}

	var team model.Team
	if err := dao.DB.First(&team, "id = ?", teamUser.TeamID).Error; err != nil {
		c.Next(err)
		return
	}

	c.JSON(apiio.GetTeamResponse{
		Response: apiio.Response{
			Success: true,
		},
		Data: struct {
			Team     model.Team
			TeamUser model.TeamUser
		}{
			Team:     team,
			TeamUser: teamUser,
		},
	})
}

// CreateTeam ..
func CreateTeam(c *fiber.Ctx) {
	user := c.Locals("user").(model.User)
	var req apiio.NewTeamRequrest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}
	if err := validator.Validator.Struct(req); err != nil {
		c.Next(err)
		return
	}

	tx := dao.DB.Begin()

	var team model.Team
	team.Name = req.Name
	team.Pubkey = req.Pubkey
	if err := tx.Save(&team).Error; err != nil {
		tx.Rollback()
		c.Next(err)
		return
	}

	var om model.TeamUser
	om.TeamID = team.ID
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
		Message: fmt.Sprintf("Add server successful %s(%d)", req.Name, team.ID),
	})
}
