package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/facades"
	"goravel/app/services"
)

type WarehouseTransferController struct {
	service      *services.WarehouseTransferService
	queryService *services.WarehouseTransferQueryService
}

func NewWarehouseTransferController(service *services.WarehouseTransferService, queryService *services.WarehouseTransferQueryService) *WarehouseTransferController {
	return &WarehouseTransferController{service: service, queryService: queryService}
}

func (c *WarehouseTransferController) Index(ctx http.Context) http.Response {
	transfers, err := c.queryService.FindAll()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "warehouse transfers retrieved successfully",
		"data":    transfers,
	})
}

func (c *WarehouseTransferController) Show(ctx http.Context) http.Response {
	transferID := ctx.Request().RouteInt64("id")
	if transferID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid warehouse transfer ID")
	}

	transfer, err := c.queryService.FindByID(uint64(transferID))
	if err != nil {
		switch {
		case errors.Is(err, frameworkerrors.OrmRecordNotFound):
			return errorResponse(ctx, http.StatusNotFound, "warehouse transfer not found")
		case errors.Is(err, services.ErrWarehouseTransferQueryInvalidID):
			return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "warehouse transfer retrieved successfully",
		"data":    transfer,
	})
}

func (c *WarehouseTransferController) Store(ctx http.Context) http.Response {
	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var input services.CreateWarehouseTransferDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}

	draft, err := c.service.CreateDraft(ctx, input, actorID)
	if err != nil {
		return mapWarehouseTransferError(ctx, err)
	}

	return ctx.Response().Json(http.StatusCreated, http.Json{
		"success": true,
		"message": "warehouse transfer draft created successfully",
		"data":    draft,
	})
}

func (c *WarehouseTransferController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid warehouse transfer ID")
	}
	var input services.CreateWarehouseTransferDraftInput
	if err := ctx.Request().Bind(&input); err != nil {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid request payload")
	}
	draft, err := c.service.UpdateDraft(ctx, uint64(id), input)
	if err != nil {
		return mapWarehouseTransferError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "warehouse transfer draft updated successfully", "data": draft})
}

func (c *WarehouseTransferController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid warehouse transfer ID")
	}
	if err := c.service.DeleteDraft(ctx, uint64(id)); err != nil {
		return mapWarehouseTransferError(ctx, err)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"success": true, "message": "warehouse transfer draft deleted successfully", "data": http.Json{}})
}

func (c *WarehouseTransferController) Post(ctx http.Context) http.Response {
	transferID := ctx.Request().RouteInt64("id")
	if transferID <= 0 {
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid warehouse transfer ID")
	}

	actorIDString, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	actorID, err := strconv.ParseUint(actorIDString, 10, 64)
	if err != nil || actorID == 0 {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := c.service.Post(ctx, uint64(transferID), actorID); err != nil {
		return mapWarehouseTransferError(ctx, err)
	}

	return ctx.Response().Json(http.StatusOK, http.Json{
		"success": true,
		"message": "warehouse transfer posted successfully",
		"data":    http.Json{},
	})
}

func mapWarehouseTransferError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, frameworkerrors.OrmRecordNotFound):
		return errorResponse(ctx, http.StatusNotFound, "warehouse transfer not found")
	case errors.Is(err, services.ErrWarehouseTransferSourceWarehouseNotFound),
		errors.Is(err, services.ErrWarehouseTransferDestinationWarehouseNotFound),
		errors.Is(err, services.ErrWarehouseTransferProductNotFound):
		return errorResponse(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrWarehouseTransferAlreadyPosted),
		errors.Is(err, services.ErrWarehouseTransferInvalidStatus),
		errors.Is(err, services.ErrWarehouseTransferInsufficientStock),
		errors.Is(err, services.ErrWarehouseTransferReferenceCollision):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrWarehouseTransferSameWarehouse),
		errors.Is(err, services.ErrWarehouseTransferEmptyItems),
		errors.Is(err, services.ErrWarehouseTransferInvalidProduct),
		errors.Is(err, services.ErrWarehouseTransferInvalidQuantity),
		errors.Is(err, services.ErrWarehouseTransferDuplicateItem),
		errors.Is(err, services.ErrWarehouseTransferNil),
		errors.Is(err, services.ErrWarehouseTransferSourceWarehouseInactive),
		errors.Is(err, services.ErrWarehouseTransferDestinationWarehouseInactive),
		errors.Is(err, services.ErrWarehouseTransferProductInactive),
		errors.Is(err, services.ErrWarehouseTransferInvalidDate),
		errors.Is(err, services.ErrWarehouseTransferMissingSourceWarehouseID),
		errors.Is(err, services.ErrWarehouseTransferMissingDestinationWarehouseID):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
