package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/services"
)

type WarehouseStockController struct {
	service *services.WarehouseStockQueryService
}

func NewWarehouseStockController(service *services.WarehouseStockQueryService) *WarehouseStockController {
	return &WarehouseStockController{service: service}
}

func (c *WarehouseStockController) Index(ctx http.Context) http.Response {
	warehouseID := ctx.Request().RouteInt64("warehouse_id")
	if warehouseID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid warehouse ID")
	}

	stocks, err := c.service.FindByWarehouse(uint64(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, frameworkerrors.OrmRecordNotFound):
			return errorResponse(ctx, http.StatusNotFound, "warehouse not found")
		case errors.Is(err, services.ErrWarehouseStockInvalidWarehouseID):
			return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "warehouse stocks retrieved successfully",
		"data":    stocks,
	})
}
