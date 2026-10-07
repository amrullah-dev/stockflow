package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/facades"
	"goravel/app/services"
)

type CategoryController struct{ service *services.CategoryService }

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

func (c *CategoryController) Index(ctx http.Context) http.Response {
	rows, err := c.service.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "categories retrieved successfully", "data": rows})
}
func (c *CategoryController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid category ID")
	}
	row, err := c.service.FindByID(uint64(id))
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "category not found")
		}
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "category retrieved successfully", "data": row})
}
func (c *CategoryController) Store(ctx http.Context) http.Response {
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var input services.CategoryInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	row, err := c.service.Create(ctx, input, actorID)
	if err != nil {
		return mapCategoryError(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{"success": true, "message": "category created successfully", "data": row})
}
func (c *CategoryController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid category ID")
	}
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var input services.CategoryInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	row, err := c.service.Update(ctx, uint64(id), input, actorID)
	if err != nil {
		return mapCategoryError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "category updated successfully", "data": row})
}
func (c *CategoryController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid category ID")
	}
	actorID, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err := c.service.Delete(ctx, uint64(id), actorID); err != nil {
		return mapCategoryError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "category deleted successfully", "data": http.Json{}})
}
func categoryActorID(ctx http.Context) (uint64, bool) {
	value, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return 0, false
	}
	id, err := strconv.ParseUint(value, 10, 64)
	return id, err == nil && id > 0
}
func mapCategoryError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, http.StatusNotFound, "category not found")
	case errors.Is(err, services.ErrCategoryDuplicate):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrCategoryInvalidID), errors.Is(err, services.ErrCategoryNameRequired):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
