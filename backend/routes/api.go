package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
	"goravel/app/repositories"
	"goravel/app/services"
)

func API() {
	stockInService := services.NewStockInService(facades.Orm())
	stockOutService := services.NewStockOutService(facades.Orm())
	warehouseTransferService := services.NewWarehouseTransferService(facades.Orm())
	query := facades.Orm().Query()
	warehouseStockQueryService := services.NewWarehouseStockQueryService(
		repositories.NewWarehouseRepository(query),
		repositories.NewWarehouseStockRepository(query),
	)
	warehouseStockController := controllers.NewWarehouseStockController(warehouseStockQueryService)
	productQueryService := services.NewProductQueryService(repositories.NewProductRepository(query))
	productController := controllers.NewProductController(productQueryService)
	productWriteController := controllers.NewProductWriteController(services.NewProductService(facades.Orm()))
	warehouseQueryService := services.NewWarehouseQueryService(repositories.NewWarehouseRepository(query))
	warehouseController := controllers.NewWarehouseController(warehouseQueryService)
	warehouseWriteController := controllers.NewWarehouseWriteController(services.NewWarehouseService(facades.Orm()))
	warehouseTransferQueryService := services.NewWarehouseTransferQueryService(repositories.NewWarehouseTransferRepository(query))
	warehouseTransferController := controllers.NewWarehouseTransferController(warehouseTransferService, warehouseTransferQueryService)
	auditLogQueryService := services.NewAuditLogQueryService(repositories.NewAuditLogRepository(query))
	auditLogController := controllers.NewAuditLogController(auditLogQueryService)
	stockInQueryService := services.NewStockInQueryService(repositories.NewStockInRepository(query))
	stockInController := controllers.NewStockInController(stockInService, stockInQueryService)
	stockOutQueryService := services.NewStockOutQueryService(repositories.NewStockOutRepository(query))
	stockOutController := controllers.NewStockOutController(stockOutService, stockOutQueryService)
	stockMovementQueryService := services.NewStockMovementQueryService(repositories.NewStockMovementRepository(query))
	stockMovementController := controllers.NewStockMovementController(stockMovementQueryService)
	categoryController := controllers.NewCategoryController(services.NewCategoryService(facades.Orm()))
	supplierController := controllers.NewSupplierController(services.NewSupplierService(facades.Orm()))
	authController := controllers.NewAuthController(services.NewAuthService(facades.Orm()))
	dashboardController := controllers.NewDashboardController(services.NewDashboardSummaryService(query))

	facades.Route().Prefix("/api").Post("/login", authController.Login)
	facades.Route().
		Prefix("/api").
		Middleware(&middleware.Authenticate{}).
		Group(func(router route.Router) {
			router.Post("/stock-ins", stockInController.Store)
			router.Put("/stock-ins/{id}", stockInController.Update)
			router.Delete("/stock-ins/{id}", stockInController.Destroy)
			router.Post("/stock-ins/{id}/post", stockInController.Post)
			router.Get("/stock-ins", stockInController.Index)
			router.Get("/stock-ins/{id}", stockInController.Show)
			router.Post("/stock-outs", stockOutController.Store)
			router.Put("/stock-outs/{id}", stockOutController.Update)
			router.Delete("/stock-outs/{id}", stockOutController.Destroy)
			router.Post("/stock-outs/{id}/post", stockOutController.Post)
			router.Get("/stock-outs", stockOutController.Index)
			router.Get("/stock-outs/{id}", stockOutController.Show)
			router.Post("/warehouse-transfers", warehouseTransferController.Store)
			router.Put("/warehouse-transfers/{id}", warehouseTransferController.Update)
			router.Delete("/warehouse-transfers/{id}", warehouseTransferController.Destroy)
			router.Post("/warehouse-transfers/{id}/post", warehouseTransferController.Post)
			router.Get("/warehouse-transfers", warehouseTransferController.Index)
			router.Get("/warehouse-transfers/{id}", warehouseTransferController.Show)
			router.Get("/dashboard/summary", dashboardController.Summary)
			router.Get("/audit-logs", auditLogController.Index)
			router.Get("/audit-logs/{id}", auditLogController.Show)
			router.Get("/stock-movements", stockMovementController.Index)
			router.Get("/warehouses/{warehouse_id}/stocks", warehouseStockController.Index)
			router.Get("/products", productController.Index)
			router.Post("/products", productWriteController.Store)
			router.Put("/products/{id}", productWriteController.Update)
			router.Delete("/products/{id}", productWriteController.Destroy)
			router.Get("/warehouses", warehouseController.Index)
			router.Post("/warehouses", warehouseWriteController.Store)
			router.Put("/warehouses/{id}", warehouseWriteController.Update)
			router.Delete("/warehouses/{id}", warehouseWriteController.Destroy)
			router.Get("/categories", categoryController.Index)
			router.Get("/categories/{id}", categoryController.Show)
			router.Post("/categories", categoryController.Store)
			router.Put("/categories/{id}", categoryController.Update)
			router.Delete("/categories/{id}", categoryController.Destroy)
			router.Get("/suppliers", supplierController.Index)
			router.Get("/suppliers/{id}", supplierController.Show)
			router.Post("/suppliers", supplierController.Store)
			router.Put("/suppliers/{id}", supplierController.Update)
			router.Delete("/suppliers/{id}", supplierController.Destroy)
		})
}
