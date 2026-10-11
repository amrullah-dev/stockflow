package controllers

import (
	"errors"
	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"
	"goravel/app/services"
)

type ProductWriteController struct{ service *services.ProductService }

func NewProductWriteController(s *services.ProductService) *ProductWriteController {
	return &ProductWriteController{service: s}
}
func (c *ProductWriteController) Store(ctx http.Context) http.Response { return c.write(ctx, 0, true) }
func (c *ProductWriteController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, 422, "invalid product ID")
	}
	return c.write(ctx, uint64(id), false)
}
func (c *ProductWriteController) write(ctx http.Context, id uint64, create bool) http.Response {
	actor, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, 401, "unauthorized")
	}
	var in services.ProductInput
	if err := ctx.Request().Bind(&in); err != nil {
		return errorResponse(ctx, 422, "invalid request payload")
	}
	var row *services.ProductResponse
	var err error
	if create {
		row, err = c.service.Create(ctx, in, actor)
	} else {
		row, err = c.service.Update(ctx, id, in, actor)
	}
	if err != nil {
		return productError(ctx, err)
	}
	status := 200
	message := "product updated successfully"
	if create {
		status = 201
		message = "product created successfully"
	}
	return ctx.Response().Json(status, http.Json{"success": true, "message": message, "data": row})
}
func (c *ProductWriteController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, 422, "invalid product ID")
	}
	actor, ok := categoryActorID(ctx)
	if !ok {
		return errorResponse(ctx, 401, "unauthorized")
	}
	if err := c.service.Delete(ctx, uint64(id), actor); err != nil {
		return productError(ctx, err)
	}
	return ctx.Response().Json(200, http.Json{"success": true, "message": "product deleted successfully", "data": http.Json{}})
}
func productError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, 404, "product or category not found")
	case errors.Is(err, services.ErrProductDuplicateCode):
		return errorResponse(ctx, 409, err.Error())
	case errors.Is(err, services.ErrProductInUse):
		return errorResponse(ctx, 409, err.Error())
	case errors.Is(err, services.ErrProductInvalidID), errors.Is(err, services.ErrProductCategoryRequired), errors.Is(err, services.ErrProductCategoryNotFound), errors.Is(err, services.ErrProductCodeRequired), errors.Is(err, services.ErrProductNameRequired), errors.Is(err, services.ErrProductUnitRequired), errors.Is(err, services.ErrProductMinimumStockInvalid):
		return errorResponse(ctx, 422, err.Error())
	default:
		return errorResponse(ctx, 500, "internal server error")
	}
}
