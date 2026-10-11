package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"
	"goravel/app/models"
)

type DashboardRecentTransaction struct {
	ID            uint64                         `json:"id"`
	Type          string                         `json:"type"`
	Quantity      int64                          `json:"quantity"`
	Warehouse     StockMovementWarehouseResponse `json:"warehouse"`
	Product       StockMovementProductResponse   `json:"product"`
	ReferenceType string                         `json:"reference_type"`
	ReferenceID   *uint64                        `json:"reference_id"`
	BalanceAfter  int64                          `json:"balance_after"`
	CreatedAt     *carbon.DateTime               `json:"created_at"`
}

type DashboardSummaryResponse struct {
	TotalProducts      int64                        `json:"total_products"`
	TotalWarehouses    int64                        `json:"total_warehouses"`
	TotalSuppliers     int64                        `json:"total_suppliers"`
	TotalStockQuantity int64                        `json:"total_stock_quantity"`
	LowStockCount      int64                        `json:"low_stock_count"`
	RecentTransactions []DashboardRecentTransaction `json:"recent_transactions"`
}

type DashboardSummaryService struct{ query orm.Query }

func NewDashboardSummaryService(query orm.Query) *DashboardSummaryService {
	return &DashboardSummaryService{query: query}
}

func (s *DashboardSummaryService) Summary() (*DashboardSummaryResponse, error) {
	products, err := s.query.Model(&models.Product{}).Where("is_active = ?", true).Count()
	if err != nil {
		return nil, err
	}
	warehouses, err := s.query.Model(&models.Warehouse{}).Where("is_active = ?", true).Count()
	if err != nil {
		return nil, err
	}
	suppliers, err := s.query.Model(&models.Supplier{}).Where("is_active = ?", true).Count()
	if err != nil {
		return nil, err
	}
	var totals struct {
		Total int64 `gorm:"column:total"`
		Low   int64 `gorm:"column:low"`
	}
	err = s.query.Table("warehouse_stocks ws").Join("JOIN products p ON p.id = ws.product_id AND p.deleted_at IS NULL").Join("JOIN warehouses w ON w.id = ws.warehouse_id AND w.deleted_at IS NULL").SelectRaw("COALESCE(SUM(ws.quantity), 0) AS total, COALESCE(SUM(CASE WHEN ws.quantity <= p.minimum_stock THEN 1 ELSE 0 END), 0) AS low").Where("p.is_active = ? AND w.is_active = ?", true, true).Scan(&totals)
	if err != nil {
		return nil, err
	}
	var rows []models.StockMovement
	err = s.query.Model(&models.StockMovement{}).With("Warehouse").With("Product").OrderBy("created_at", "desc").OrderBy("id", "desc").Limit(10).Find(&rows)
	if err != nil {
		return nil, err
	}
	recent := make([]DashboardRecentTransaction, 0, len(rows))
	for _, row := range rows {
		recent = append(recent, DashboardRecentTransaction{ID: uint64(row.ID), Type: row.Type, Quantity: row.Quantity, Warehouse: StockMovementWarehouseResponse{ID: uint64(row.Warehouse.ID), Code: row.Warehouse.Code, Name: row.Warehouse.Name}, Product: StockMovementProductResponse{ID: uint64(row.Product.ID), Code: row.Product.Code, Name: row.Product.Name, Unit: row.Product.Unit}, ReferenceType: row.ReferenceType, ReferenceID: row.ReferenceID, BalanceAfter: row.BalanceAfter, CreatedAt: row.CreatedAt})
	}
	return &DashboardSummaryResponse{TotalProducts: products, TotalWarehouses: warehouses, TotalSuppliers: suppliers, TotalStockQuantity: totals.Total, LowStockCount: totals.Low, RecentTransactions: recent}, nil
}
