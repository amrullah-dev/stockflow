package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type DashboardController struct {
}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

func (c *DashboardController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("dashboard.tmpl", map[string]any{
		"title": "Inventory Dashboard",
	})
}