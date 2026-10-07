package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/services"
)

type WarehouseController struct {
	service *services.WarehouseQueryService
}

func NewWarehouseController(service *services.WarehouseQueryService) *WarehouseController {
	return &WarehouseController{service: service}
}

func (c *WarehouseController) Index(ctx http.Context) http.Response {
	warehouses, err := c.service.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "warehouses retrieved successfully",
		"data":    warehouses,
	})
}
