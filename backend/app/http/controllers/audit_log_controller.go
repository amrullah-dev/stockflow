package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/services"
)

type AuditLogController struct {
	queryService *services.AuditLogQueryService
}

func NewAuditLogController(queryService *services.AuditLogQueryService) *AuditLogController {
	return &AuditLogController{queryService: queryService}
}

func (c *AuditLogController) Index(ctx http.Context) http.Response {
	logs, err := c.queryService.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "audit logs retrieved successfully",
		"data":    logs,
	})
}

func (c *AuditLogController) Show(ctx http.Context) http.Response {
	logID := ctx.Request().RouteInt64("id")
	if logID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid audit log ID")
	}

	log, err := c.queryService.FindByID(uint64(logID))
	if err != nil {
		switch {
		case errors.Is(err, frameworkerrors.OrmRecordNotFound):
			return errorResponse(ctx, http.StatusNotFound, "audit log not found")
		case errors.Is(err, services.ErrAuditLogQueryInvalidID):
			return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "audit log retrieved successfully",
		"data":    log,
	})
}
