package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/services"
)

type AuthController struct{ service *services.AuthService }

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

func (c *AuthController) Login(ctx http.Context) http.Response {
	var input services.LoginInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "validation failed")
	}
	result, err := c.service.Login(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrLoginEmailRequired), errors.Is(err, services.ErrLoginPasswordRequired), errors.Is(err, services.ErrLoginEmailInvalid):
			return errorResponse(ctx, http.StatusUnprocessableEntity, "validation failed")
		case errors.Is(err, services.ErrLoginInvalidCredentials):
			return errorResponse(ctx, http.StatusUnauthorized, "invalid credentials")
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "login successful",
		"data":    result,
	})
}
