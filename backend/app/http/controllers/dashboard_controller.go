package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"goravel/app/services"
)

type DashboardController struct {
	service *services.DashboardSummaryService
}

func NewDashboardController(service ...*services.DashboardSummaryService) *DashboardController {
	c := &DashboardController{}
	if len(service) > 0 {
		c.service = service[0]
	}
	return c
}
func (c *DashboardController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("dashboard.tmpl", map[string]any{"title": "Inventory Dashboard"})
}
func (c *DashboardController) Summary(ctx http.Context) http.Response {
	if c.service == nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	summary, err := c.service.Summary()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "dashboard summary retrieved successfully", "data": summary})
}
