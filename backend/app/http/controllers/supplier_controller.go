package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/services"
)

type SupplierController struct{ service *services.SupplierService }

func NewSupplierController(service *services.SupplierService) *SupplierController {
	return &SupplierController{service: service}
}
func (c *SupplierController) Index(ctx http.Context) http.Response {
	rows, err := c.service.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "suppliers retrieved successfully", "data": rows})
}
func (c *SupplierController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid supplier ID")
	}
	row, err := c.service.FindByID(uint64(id))
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "supplier not found")
		}
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "supplier retrieved successfully", "data": row})
}
func (c *SupplierController) Store(ctx http.Context) http.Response {
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var input services.SupplierInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	row, err := c.service.Create(ctx, input, actorID)
	if err != nil {
		return mapSupplierError(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{"success": true, "message": "supplier created successfully", "data": row})
}
func (c *SupplierController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid supplier ID")
	}
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var input services.SupplierInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	row, err := c.service.Update(ctx, uint64(id), input, actorID)
	if err != nil {
		return mapSupplierError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "supplier updated successfully", "data": row})
}
func (c *SupplierController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid supplier ID")
	}
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err := c.service.Delete(ctx, uint64(id), actorID); err != nil {
		return mapSupplierError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "supplier deleted successfully", "data": http.Json{}})
}
func mapSupplierError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, http.StatusNotFound, "supplier not found")
	case errors.Is(err, services.ErrSupplierDuplicateCode):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrSupplierInvalidID), errors.Is(err, services.ErrSupplierCodeRequired), errors.Is(err, services.ErrSupplierNameRequired):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
