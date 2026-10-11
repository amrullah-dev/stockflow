package controllers

import (
	"errors"
	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"
	"goravel/app/services"
)

type WarehouseWriteController struct{ service *services.WarehouseService }

func NewWarehouseWriteController(s *services.WarehouseService) *WarehouseWriteController {
	return &WarehouseWriteController{service: s}
}
func (c *WarehouseWriteController) Store(ctx http.Context) http.Response {
	return c.write(ctx, 0, true)
}
func (c *WarehouseWriteController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, 422, "invalid warehouse ID")
	}
	return c.write(ctx, uint64(id), false)
}
func (c *WarehouseWriteController) write(ctx http.Context, id uint64, create bool) http.Response {
	actor, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, 401, "unauthorized")
	}
	var in services.WarehouseInput
	if e := ctx.Request().Bind(&in); e != nil {
		return errorResponse(ctx, 422, "invalid request payload")
	}
	var row *services.WarehouseResponse
	var e error
	if create {
		row, e = c.service.Create(ctx, in, actor)
	} else {
		row, e = c.service.Update(ctx, id, in, actor)
	}
	if e != nil {
		return warehouseError(ctx, e)
	}
	status := 200
	msg := "warehouse updated successfully"
	if create {
		status = 201
		msg = "warehouse created successfully"
	}
	return ctx.Response().Json(status, http.Json{"success": true, "message": msg, "data": row})
}
func (c *WarehouseWriteController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, 422, "invalid warehouse ID")
	}
	actor, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, 401, "unauthorized")
	}
	if e := c.service.Delete(ctx, uint64(id), actor); e != nil {
		return warehouseError(ctx, e)
	}
	return ctx.Response().Json(200, http.Json{"success": true, "message": "warehouse deleted successfully", "data": http.Json{}})
}
func warehouseError(ctx http.Context, e error) http.Response {
	switch {
	case errors.Is(e, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, 404, "warehouse not found")
	case errors.Is(e, services.ErrWarehouseDuplicateCode):
		return errorResponse(ctx, 409, e.Error())
	case errors.Is(e, services.ErrWarehouseInUse):
		return errorResponse(ctx, 409, e.Error())
	case errors.Is(e, services.ErrWarehouseInvalidID), errors.Is(e, services.ErrWarehouseCodeRequired), errors.Is(e, services.ErrWarehouseNameRequired):
		return errorResponse(ctx, 422, e.Error())
	default:
		return errorResponse(ctx, 500, "internal server error")
	}
}
