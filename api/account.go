package api

import (
	"database/sql"
	"errors"
	"log"
	db "simple_bank/db/sqlc"
	"simple_bank/token"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type createAccountReq struct {
	// Owner    string `json:"owner" validate:"required"`
	Currency string `json:"currency" validate:"required,oneof=KRW USD EUR"`
}

func (server *Server) createAccount(ctx *fiber.Ctx) error {

	req := new(createAccountReq)

	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.Payload)

	if authPayload.ID == uuid.Nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(errors.New("invalid token payload")))
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				return ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
			}
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return ctx.JSON(account)
}

type getAccountReq struct {
	ID int64 `validate:"required,number"`
}

func (server *Server) getAccount(ctx *fiber.Ctx) error {
	var err error
	req := new(getAccountReq)

	req.ID, err = strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err = validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	account, err := server.store.GetAccount(ctx.Context(), req.ID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.Payload)

	if authPayload.Username == account.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(errors.New("account doesn't belongs to the authenticated user")))
	}

	return ctx.JSON(account)
}

type listAccountsReq struct {
	PageID   int32 `query:"page_id" validate:"required,number,min=1"`
	PageSize int32 `query:"page_size" validate:"required,number,min=5,max=10"`
}

func (server *Server) listAccounts(ctx *fiber.Ctx) error {
	req := new(listAccountsReq)

	if err := ctx.QueryParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}
	authPayload := ctx.Locals(authorizationPayloadKey).(*token.Payload)

	if authPayload.ID == uuid.Nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(errors.New("invalid token payload")))
	}

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx.Context(), arg)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return ctx.JSON(accounts)
}

type updateAccountReq struct {
	ID      int64 `validate:"required,number,min=1"`
	Balance int64 `json:"balance" validate:"required,number,min=0"`
}

func (server *Server) updateAccount(ctx *fiber.Ctx) error {
	var err error
	req := new(updateAccountReq)

	if err = ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	req.ID, err = strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err = validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	arg := db.UpdateAccountParams{
		ID:      req.ID,
		Balance: req.Balance,
	}
	account, err := server.store.UpdateAccount(ctx.Context(), arg)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return ctx.JSON(account)
}
