package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/facades"
	"goravel/app/services"
)

type StockInController struct {
	service      *services.StockInService
	queryService *services.StockInQueryService
}

func NewStockInController(service *services.StockInService, queryService *services.StockInQueryService) *StockInController {
	return &StockInController{service: service, queryService: queryService}
}

func (c *StockInController) Index(ctx http.Context) http.Response {
	stockIns, err := c.queryService.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock-ins retrieved successfully",
		"data":    stockIns,
	})
}

func (c *StockInController) Show(ctx http.Context) http.Response {
	stockInID := ctx.Request().RouteInt64("id")
	if stockInID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-in ID")
	}

	stockIn, err := c.queryService.FindByID(uint64(stockInID))
	if err != nil {
		switch {
		case errors.Is(err, frameworkerrors.OrmRecordNotFound):
			return errorResponse(ctx, http.StatusNotFound, "stock-in not found")
		case errors.Is(err, services.ErrStockInQueryInvalidID):
			return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock-in retrieved successfully",
		"data":    stockIn,
	})
}

func (c *StockInController) Store(ctx http.Context) http.Response {
	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var input services.CreateStockInDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}

	draft, err := c.service.CreateDraft(ctx, input, actorID)
	if err != nil {
		return mapStockInError(ctx, err)
	}

	return ctx.Response().Json(http.StatusCreated, http.Json{
		"success": true,
		"message": "stock in draft created successfully",
		"data":    draft,
	})
}

func (c *StockInController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-in ID")
	}
	var input services.CreateStockInDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	draft, err := c.service.UpdateDraft(ctx, uint64(id), input)
	if err != nil {
		return mapStockInError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "stock in draft updated successfully", "data": draft})
}

func (c *StockInController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-in ID")
	}
	if err := c.service.DeleteDraft(ctx, uint64(id)); err != nil {
		return mapStockInError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "stock in draft deleted successfully", "data": http.Json{}})
}

func (c *StockInController) Post(ctx http.Context) http.Response {
	stockInID := ctx.Request().RouteInt64("id")
	if stockInID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-in ID")
	}

	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := c.service.Post(ctx, uint64(stockInID), actorID); err != nil {
		return mapStockInError(ctx, err)
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock in posted successfully",
		"data":    http.Json{},
	})
}

func mapStockInError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, http.StatusNotFound, "stock-in not found")
	case errors.Is(err, services.ErrStockInSupplierNotFound),
		errors.Is(err, services.ErrStockInWarehouseNotFound),
		errors.Is(err, services.ErrStockInProductNotFound):
		return errorResponse(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrStockInAlreadyPosted),
		errors.Is(err, services.ErrStockInInvalidStatus),
		errors.Is(err, services.ErrStockInReferenceCollision):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrStockInEmptyItems),
		errors.Is(err, services.ErrStockInInvalidProduct),
		errors.Is(err, services.ErrStockInInvalidQty),
		errors.Is(err, services.ErrStockInDuplicateItem),
		errors.Is(err, services.ErrStockInNil),
		errors.Is(err, services.ErrStockInSupplierInactive),
		errors.Is(err, services.ErrStockInWarehouseInactive),
		errors.Is(err, services.ErrStockInProductInactive),
		errors.Is(err, services.ErrStockInInvalidDate),
		errors.Is(err, services.ErrStockInMissingSupplierID),
		errors.Is(err, services.ErrStockInMissingWarehouseID),
		errors.Is(err, services.ErrStockInInvalidUnitPrice):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}

func errorResponse(ctx http.Context, status int, message string) http.Response {
	return ctx.Response().Json(status, http.Json{
		"success": false,
		"message": message,
		"errors":  http.Json{},
	})
}
