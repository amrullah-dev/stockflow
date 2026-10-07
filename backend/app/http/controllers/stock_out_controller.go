package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/facades"
	"goravel/app/services"
)

type StockOutController struct {
	service      *services.StockOutService
	queryService *services.StockOutQueryService
}

func NewStockOutController(service *services.StockOutService, queryService *services.StockOutQueryService) *StockOutController {
	return &StockOutController{service: service, queryService: queryService}
}

func (c *StockOutController) Index(ctx http.Context) http.Response {
	stockOuts, err := c.queryService.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock-outs retrieved successfully",
		"data":    stockOuts,
	})
}

func (c *StockOutController) Show(ctx http.Context) http.Response {
	stockOutID := ctx.Request().RouteInt64("id")
	if stockOutID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-out ID")
	}

	stockOut, err := c.queryService.FindByID(uint64(stockOutID))
	if err != nil {
		switch {
		case errors.Is(err, frameworkerrors.OrmRecordNotFound):
			return errorResponse(ctx, http.StatusNotFound, "stock-out not found")
		case errors.Is(err, services.ErrStockOutQueryInvalidID):
			return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock-out retrieved successfully",
		"data":    stockOut,
	})
}

func (c *StockOutController) Store(ctx http.Context) http.Response {
	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var input services.CreateStockOutDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}

	draft, err := c.service.CreateDraft(ctx, input, actorID)
	if err != nil {
		return mapStockOutError(ctx, err)
	}

	return ctx.Response().Json(http.StatusCreated, http.Json{
		"success": true,
		"message": "stock out draft created successfully",
		"data":    draft,
	})
}

func (c *StockOutController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-out ID")
	}
	var input services.CreateStockOutDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	draft, err := c.service.UpdateDraft(ctx, uint64(id), input)
	if err != nil {
		return mapStockOutError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "stock out draft updated successfully", "data": draft})
}

func (c *StockOutController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-out ID")
	}
	if err := c.service.DeleteDraft(ctx, uint64(id)); err != nil {
		return mapStockOutError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "stock out draft deleted successfully", "data": http.Json{}})
}

func (c *StockOutController) Post(ctx http.Context) http.Response {
	stockOutID := ctx.Request().RouteInt64("id")
	if stockOutID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid stock-out ID")
	}

	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := c.service.Post(ctx, uint64(stockOutID), actorID); err != nil {
		return mapStockOutError(ctx, err)
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "stock out posted successfully",
		"data":    http.Json{},
	})
}

func mapStockOutError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, http.StatusNotFound, "stock-out not found")
	case errors.Is(err, services.ErrStockOutWarehouseNotFound),
		errors.Is(err, services.ErrStockOutProductNotFound):
		return errorResponse(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrStockOutAlreadyPosted),
		errors.Is(err, services.ErrStockOutInvalidStatus),
		errors.Is(err, services.ErrStockOutInsufficientStock),
		errors.Is(err, services.ErrStockOutReferenceCollision):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrStockOutEmptyItems),
		errors.Is(err, services.ErrStockOutInvalidProduct),
		errors.Is(err, services.ErrStockOutInvalidQty),
		errors.Is(err, services.ErrStockOutDuplicateItem),
		errors.Is(err, services.ErrStockOutNil),
		errors.Is(err, services.ErrStockOutWarehouseInactive),
		errors.Is(err, services.ErrStockOutProductInactive),
		errors.Is(err, services.ErrStockOutInvalidDate),
		errors.Is(err, services.ErrStockOutMissingWarehouseID):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
