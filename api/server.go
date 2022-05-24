package api

import (
	db "simple_bank/db/sqlc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	store  *db.Store
	router *fiber.App
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := fiber.New()

	router.Use(logger.New())
	router.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON("pong")
	})

	router.Post("/accounts", server.createAccount)
	router.Get("/account/:id", server.getAccount)
	router.Get("/accounts", server.listAccounts)
	router.Put("/account/:id", server.updateAccount)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Listen(address)
}

func errorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"error": err.Error(),
	}
}
