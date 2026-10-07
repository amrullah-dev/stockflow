package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/services"
)

type StockMovementController struct {
	queryService *services.StockMovementQueryService
}

func NewStockMovementController(queryService *services.StockMovementQueryService) *StockMovementController {
	return &StockMovementController{queryService: queryService}
}

func (c *StockMovementController) Index(ctx http.Context) http.Response {
	movements, err := c.queryService.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock movements retrieved successfully",
		"data":    movements,
	})
}
