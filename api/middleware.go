package api

import (
	"errors"
	"fmt"
	"simple_bank/token"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		authorizationHeader := ctx.GetReqHeaders()[authorizationHeaderKey]
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid athorization header format")
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		}
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type: %s", authorizationType)
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		}

		ctx.Locals(authorizationPayloadKey, payload)
		return ctx.Next()
	}
}
