package middleware

import (
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type Authenticate struct{}

func (m *Authenticate) Signature() string {
	return "auth"
}

func (m *Authenticate) Handle(ctx contractshttp.Context) {
	authorization := strings.TrimSpace(ctx.Request().Header("Authorization"))
	token, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		unauthorized(ctx)
		return
	}

	if !authenticate(ctx, strings.TrimSpace(token)) {
		unauthorized(ctx)
		return
	}

	ctx.Request().Next()
}

func authenticate(ctx contractshttp.Context, token string) (authenticated bool) {
	defer func() {
		if recover() != nil {
			authenticated = false
		}
	}()

	_, err := facades.Auth(ctx).Guard("user").Parse(token)
	return err == nil
}

func unauthorized(ctx contractshttp.Context) {
	_ = ctx.Response().Json(contractshttp.StatusUnauthorized, contractshttp.Json{
		"success": false,
		"message": "unauthorized",
	}).Abort()
}
