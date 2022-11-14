package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "simple_bank/db/sqlc"
	"simple_bank/token"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" validate:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" validate:"required,min=1"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	Currency      string `json:"currency" validate:"required,oneof=KRW USD EUR"`
}

func (server *Server) createTransfer(ctx *fiber.Ctx) error {
	req := new(transferRequest)
	// var req transferRequest
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}
	fromAccount, valid := server.validateAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		err := errors.New("invalid from_account currency")
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.Payload)

	if authPayload.Username == fromAccount.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(errors.New("fromAccount doesn't belongs to the authenticated user")))
	}

	_, valid = server.validateAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		err := errors.New("invalid to_account currency")
		return ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx.Context(), arg)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return ctx.JSON(result)
}

func (server *Server) validateAccount(ctx *fiber.Ctx, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx.Context(), accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			// ctx.JSON()
			return account, false
		}

		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		if err != nil {
			return account, false
		}
	}

	return account, true
}
