package api

import (
	"database/sql"
	"log"
	db "simple_bank/db/sqlc"
	"simple_bank/util"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"user_name" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreateAt          time.Time `json:"create_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CreateAt:          user.CreateAt,
		PasswordChangedAt: user.PasswordChangedAt,
	}
}

func (server *Server) createUser(ctx *fiber.Ctx) error {

	req := new(createUserRequest)

	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "unique_violation":
				return ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
			}
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	rsp := newUserResponse(user)

	return ctx.JSON(rsp)
}

type loginUserRequest struct {
	Username string `json:"user_name" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *fiber.Ctx) error {
	var req loginUserRequest

	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	user, err := server.store.GetUser(ctx.Context(), req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	return ctx.Status(fiber.StatusOK).JSON(rsp)
}
