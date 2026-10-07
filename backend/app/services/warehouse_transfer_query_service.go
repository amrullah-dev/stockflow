package services

import (
	"errors"
	"sort"
	"time"

	"goravel/app/models"
	"goravel/app/repositories"
)

var ErrWarehouseTransferQueryInvalidID = errors.New("warehouse transfer ID must be greater than zero")

type WarehouseTransferWarehouseResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type WarehouseTransferUserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WarehouseTransferProductResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type WarehouseTransferItemResponse struct {
	ID         uint64                           `json:"id"`
	TransferID uint64                           `json:"transfer_id"`
	ProductID  uint64                           `json:"product_id"`
	Quantity   int64                            `json:"quantity"`
	Product    WarehouseTransferProductResponse `json:"product"`
}

type WarehouseTransferListResponse struct {
	ID              uint64                             `json:"id"`
	ReferenceNumber string                             `json:"reference_number"`
	FromWarehouseID uint64                             `json:"from_warehouse_id"`
	ToWarehouseID   uint64                             `json:"to_warehouse_id"`
	TransactionDate time.Time                          `json:"transaction_date"`
	Status          string                             `json:"status"`
	Note            string                             `json:"note"`
	CreatedBy       uint64                             `json:"created_by"`
	PostedAt        *time.Time                         `json:"posted_at"`
	FromWarehouse   WarehouseTransferWarehouseResponse `json:"from_warehouse"`
	ToWarehouse     WarehouseTransferWarehouseResponse `json:"to_warehouse"`
}

type WarehouseTransferDetailResponse struct {
	ID              uint64                             `json:"id"`
	ReferenceNumber string                             `json:"reference_number"`
	FromWarehouseID uint64                             `json:"from_warehouse_id"`
	ToWarehouseID   uint64                             `json:"to_warehouse_id"`
	TransactionDate time.Time                          `json:"transaction_date"`
	Status          string                             `json:"status"`
	Note            string                             `json:"note"`
	CreatedBy       uint64                             `json:"created_by"`
	PostedAt        *time.Time                         `json:"posted_at"`
	FromWarehouse   WarehouseTransferWarehouseResponse `json:"from_warehouse"`
	ToWarehouse     WarehouseTransferWarehouseResponse `json:"to_warehouse"`
	CreatedByUser   WarehouseTransferUserResponse      `json:"created_by_user"`
	Items           []WarehouseTransferItemResponse    `json:"items"`
}

type WarehouseTransferQueryService struct {
	repo *repositories.WarehouseTransferRepository
}

func NewWarehouseTransferQueryService(repo *repositories.WarehouseTransferRepository) *WarehouseTransferQueryService {
	return &WarehouseTransferQueryService{repo: repo}
}

func (s *WarehouseTransferQueryService) FindAll() ([]WarehouseTransferListResponse, error) {
	transfers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]WarehouseTransferListResponse, 0, len(transfers))
	for _, transfer := range transfers {
		responses = append(responses, warehouseTransferListResponse(transfer))
	}

	return responses, nil
}

func (s *WarehouseTransferQueryService) FindByID(id uint64) (*WarehouseTransferDetailResponse, error) {
	if id == 0 {
		return nil, ErrWarehouseTransferQueryInvalidID
	}

	transfer, err := s.repo.FindByIDForRead(id)
	if err != nil {
		return nil, err
	}

	sort.Slice(transfer.Items, func(i, j int) bool {
		return transfer.Items[i].ID < transfer.Items[j].ID
	})

	response := warehouseTransferDetailResponse(*transfer)
	return &response, nil
}

func warehouseTransferListResponse(transfer models.WarehouseTransfer) WarehouseTransferListResponse {
	return WarehouseTransferListResponse{
		ID:              uint64(transfer.ID),
		ReferenceNumber: transfer.ReferenceNumber,
		FromWarehouseID: transfer.FromWarehouseID,
		ToWarehouseID:   transfer.ToWarehouseID,
		TransactionDate: transfer.TransactionDate,
		Status:          transfer.Status,
		Note:            transfer.Note,
		CreatedBy:       transfer.CreatedBy,
		PostedAt:        transfer.PostedAt,
		FromWarehouse:   warehouseTransferWarehouseResponse(transfer.FromWarehouse),
		ToWarehouse:     warehouseTransferWarehouseResponse(transfer.ToWarehouse),
	}
}

func warehouseTransferDetailResponse(transfer models.WarehouseTransfer) WarehouseTransferDetailResponse {
	items := make([]WarehouseTransferItemResponse, 0, len(transfer.Items))
	for _, item := range transfer.Items {
		items = append(items, WarehouseTransferItemResponse{
			ID:         uint64(item.ID),
			TransferID: item.TransferID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Product: WarehouseTransferProductResponse{
				ID:   uint64(item.Product.ID),
				Code: item.Product.Code,
				Name: item.Product.Name,
				Unit: item.Product.Unit,
			},
		})
	}

	return WarehouseTransferDetailResponse{
		ID:              uint64(transfer.ID),
		ReferenceNumber: transfer.ReferenceNumber,
		FromWarehouseID: transfer.FromWarehouseID,
		ToWarehouseID:   transfer.ToWarehouseID,
		TransactionDate: transfer.TransactionDate,
		Status:          transfer.Status,
		Note:            transfer.Note,
		CreatedBy:       transfer.CreatedBy,
		PostedAt:        transfer.PostedAt,
		FromWarehouse:   warehouseTransferWarehouseResponse(transfer.FromWarehouse),
		ToWarehouse:     warehouseTransferWarehouseResponse(transfer.ToWarehouse),
		CreatedByUser: WarehouseTransferUserResponse{
			ID:    uint64(transfer.CreatedByUser.ID),
			Name:  transfer.CreatedByUser.Name,
			Email: transfer.CreatedByUser.Email,
		},
		Items: items,
	}
}

func warehouseTransferWarehouseResponse(warehouse models.Warehouse) WarehouseTransferWarehouseResponse {
	return WarehouseTransferWarehouseResponse{
		ID:   uint64(warehouse.ID),
		Code: warehouse.Code,
		Name: warehouse.Name,
	}
}
