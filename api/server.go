package api

import (
	"fmt"
	db "simple_bank/db/sqlc"
	"simple_bank/token"
	"simple_bank/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	config     util.Config
	store      *db.Store
	router     *fiber.App
	tokenMaker token.Maker
}

func NewServer(config util.Config, store *db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	// router := fiber.New()

	// router.Use(logger.New())
	// router.Get("/ping", func(c *fiber.Ctx) error {
	// 	return c.JSON("pong")
	// })
	// router.Post("/users", server.createUser)
	// router.Post("/users/login", server.loginUser)

	// router.Post("/accounts", server.createAccount)
	// router.Get("/account/:id", server.getAccount)
	// router.Get("/accounts", server.listAccounts)
	// router.Put("/account/:id", server.updateAccount)
	// router.Post("/transfers", server.createTransfer)

	// server.router = router
	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := fiber.New()

	router.Use(logger.New())
	router.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON("pong")
	})
	router.Post("/users", server.createUser)
	router.Post("/users/login", server.loginUser)

	authRoutes := router.Group("/", authMiddleware(server.tokenMaker))

	authRoutes.Post("/accounts", server.createAccount)
	authRoutes.Get("/account/:id", server.getAccount)
	authRoutes.Get("/accounts", server.listAccounts)
	authRoutes.Put("/account/:id", server.updateAccount)
	authRoutes.Post("/transfers", server.createTransfer)

	// router.Post("/accounts", server.createAccount)
	// router.Get("/account/:id", server.getAccount)
	// router.Get("/accounts", server.listAccounts)
	// router.Put("/account/:id", server.updateAccount)
	// router.Post("/transfers", server.createTransfer)

	server.router = router

}

func (server *Server) Start(address string) error {
	return server.router.Listen(address)
}

func errorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"error": err.Error(),
	}
}
