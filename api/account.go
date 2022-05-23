package api

import (
	"net/http"
	db "simple_bank/db/sqlc"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type createAccountReq struct {
	Owner    string `json:"owner" validate:"required"`
	Currency string `json:"currency" validate:"required,oneof=KRW USD EUR"`
}

func (server *Server) createAccount(ctx *fiber.Ctx) error {

	req := new(createAccountReq)

	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"err": err.Error()})
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": err.Error()})
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx.Context(), arg)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"err": err.Error()})
	}

	return ctx.JSON(account)
}
