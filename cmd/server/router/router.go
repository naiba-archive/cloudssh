package router

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"

	"github.com/naiba/cloudssh/cmd/server/dao"
	"github.com/naiba/cloudssh/cmd/server/handler"
	"github.com/naiba/cloudssh/cmd/server/middleware"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
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

	app.Use(timer)
	app.Use(logger.New())
	app.Use(middleware.Auth)

	auth := app.Group("/auth")
	auth.Get("/logout", middleware.Protected, handler.Logout)
	auth.Post("/signup", handler.SignUp)
	auth.Post("/login", handler.Login)

	server := app.Group("/server", middleware.Protected)
	server.Post("/", handler.CreateServer)
	server.Post("/batch-delete", handler.BatchDelete)
	server.Patch("/:id", handler.EditServer)
	server.Get("/:id", handler.GetServer)
	server.Get("/", handler.ListServer)

	app.Listen(port)
}

func timer(c *fiber.Ctx) {
	// start timer
	start := time.Now()
	// next routes
	c.Next()
	// stop timer
	stop := time.Now()
	// Do something with response
	c.Append("Server-Timing", fmt.Sprintf("app;dur=%v", stop.Sub(start).String()))
}
