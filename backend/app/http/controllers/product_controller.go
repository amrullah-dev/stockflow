package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/services"
)

type ProductController struct {
	service *services.ProductQueryService
}

func NewProductController(service *services.ProductQueryService) *ProductController {
	return &ProductController{service: service}
}

func (c *ProductController) Index(ctx http.Context) http.Response {
	products, err := c.service.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "products retrieved successfully",
		"data":    products,
	})
}
