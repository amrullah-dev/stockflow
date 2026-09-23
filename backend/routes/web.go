package routes

import (
	"goravel/app/facades"
	"goravel/app/http/controllers"
)

func Web() {
	dashboardController := controllers.NewDashboardController()

	facades.Route().Get("/", dashboardController.Index)

	facades.Route().Static("public", "./public")

	userController := controllers.NewUserController()
	facades.Route().Get("/users", userController.Index)
}