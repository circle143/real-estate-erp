package flat

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetAllSocietyFlats struct{}

func (h *hGetAllSocietyFlats) execute(db *gorm.DB, orgId, societyRera, cursor, filter string) (*custom.PaginatedData, error) {
	var flatData []models.Flat
	query := db.
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Where("towers.society_id = ? AND towers.org_id = ?", societyRera, orgId).
		Preload("SaleDetail").
		Preload("SaleDetail.Customers").
		Preload("SaleDetail.CompanyCustomer").
		Order("flats.created_at DESC").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("flats.created_at < ?", decodedCursor)
		}
	}

	if filter == "1" || filter == "2" {
		// 1 -> sold and 2 -> unsold
		if filter == "1" {
			query = query.Where("EXISTS (SELECT 1 FROM sales WHERE sales.flat_id = flats.id)")
		} else {
			query = query.Where("NOT EXISTS (SELECT 1 FROM sales WHERE sales.flat_id = flats.id)")
		}
	}

	err := query.Find(&flatData).Error
	if err != nil {
		return nil, err
	}

	// get sale id
	var saleIDs []uuid.UUID
	for _, flat := range flatData {
		if flat.SaleDetail != nil {
			saleIDs = append(saleIDs, flat.SaleDetail.Id)
		}
	}

	// get sale amount
	var salePayments []models.SalePaid
	if len(saleIDs) > 0 {
		err := db.Model(&models.SalePaymentStatus{}).
			Select("sale_id, SUM(amount) AS total_paid_amount").
			Where("sale_id IN ?", saleIDs).
			Group("sale_id").
			Scan(&salePayments).Error
		if err != nil {
			return nil, err
		}
	}

	// map sale id -> paid amount
	totalsMap := make(map[uuid.UUID]decimal.Decimal)
	for _, sp := range salePayments {
		totalsMap[sp.SaleId] = sp.TotalPaidAmount
	}

	// add in flatData
	for i := range flatData {
		if flatData[i].SaleDetail != nil {
			flatData[i].SaleDetail.Paid = totalsMap[flatData[i].SaleDetail.Id]
			//flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice - totalsMap[flatData[i].SaleDetail.Id]
			flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice.Sub(totalsMap[flatData[i].SaleDetail.Id])
		}
	}

	return common.CreatePaginatedResponse(&flatData), nil
}

func (s *flatService) getAllSocietyFlats(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	filter := r.URL.Query().Get("filter")
	societyRera := chi.URLParam(r, "society")

	flat := hGetAllSocietyFlats{}
	res, err := flat.execute(s.db, orgId, societyRera, cursor, filter)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetAllTowerFlats struct{}

func (h *hGetAllTowerFlats) execute(db *gorm.DB, orgId, societyRera, towerId, cursor, filter string) (*custom.PaginatedData, error) {
	var flatData []models.Flat
	query := db.
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Where("flats.tower_id = ? AND towers.society_id = ? AND towers.org_id = ?", towerId, societyRera, orgId).
		Preload("SaleDetail").
		Preload("SaleDetail.Customers").
		Preload("SaleDetail.CompanyCustomer").
		Order("flats.created_at DESC").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("flats.created_at < ?", decodedCursor)
		}
	}

	if filter == "1" || filter == "2" {
		// 1 -> sold and 2 -> unsold
		if filter == "1" {
			query = query.Where("EXISTS (SELECT 1 FROM sales WHERE sales.flat_id = flats.id)")
		} else {
			query = query.Where("NOT EXISTS (SELECT 1 FROM sales WHERE sales.flat_id = flats.id)")
		}
	}

	result := query.Find(&flatData)
	if result.Error != nil {
		return nil, result.Error
	}

	// get sale id
	var saleIDs []uuid.UUID
	for _, flat := range flatData {
		if flat.SaleDetail != nil {
			saleIDs = append(saleIDs, flat.SaleDetail.Id)
		}
	}

	// get sale amount
	var salePayments []models.SalePaid
	if len(saleIDs) > 0 {
		err := db.Model(&models.SalePaymentStatus{}).
			Select("sale_id, SUM(amount) AS total_paid_amount").
			Where("sale_id IN ?", saleIDs).
			Group("sale_id").
			Scan(&salePayments).Error
		if err != nil {
			return nil, err
		}
	}

	// map sale id -> paid amount
	totalsMap := make(map[uuid.UUID]decimal.Decimal)
	for _, sp := range salePayments {
		totalsMap[sp.SaleId] = sp.TotalPaidAmount
	}

	// add in flatData
	for i := range flatData {
		if flatData[i].SaleDetail != nil {
			flatData[i].SaleDetail.Paid = totalsMap[flatData[i].SaleDetail.Id]
			//flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice - totalsMap[flatData[i].SaleDetail.Id]
			flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice.Sub(totalsMap[flatData[i].SaleDetail.Id])
		}
	}

	return common.CreatePaginatedResponse(&flatData), nil
}

func (s *flatService) getAllTowerFlats(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	filter := r.URL.Query().Get("filter")
	societyRera := chi.URLParam(r, "society")
	towerId := chi.URLParam(r, "tower")

	flat := hGetAllTowerFlats{}
	res, err := flat.execute(s.db, orgId, societyRera, towerId, cursor, filter)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetSocietyFlatByName struct{}

func (h *hGetSocietyFlatByName) execute(db *gorm.DB, orgId, society, name, cursor string) (*custom.PaginatedData, error) {
	var flatData []models.Flat
	query := db.
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Where("towers.society_id = ? AND towers.org_id = ? and flats.name like ?", society, orgId, name+"%").
		Preload("SaleDetail").
		Preload("SaleDetail.Customers").
		Preload("SaleDetail.CompanyCustomer").
		Order("flats.created_at DESC").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("flats.created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&flatData)
	if result.Error != nil {
		return nil, result.Error
	}

	// get sale id
	var saleIDs []uuid.UUID
	for _, flat := range flatData {
		if flat.SaleDetail != nil {
			saleIDs = append(saleIDs, flat.SaleDetail.Id)
		}
	}

	// get sale amount
	var salePayments []models.SalePaid
	if len(saleIDs) > 0 {
		err := db.Model(&models.SalePaymentStatus{}).
			Select("sale_id, SUM(amount) AS total_paid_amount").
			Where("sale_id IN ?", saleIDs).
			Group("sale_id").
			Scan(&salePayments).Error
		if err != nil {
			return nil, err
		}
	}

	// map sale id -> paid amount
	totalsMap := make(map[uuid.UUID]decimal.Decimal)
	for _, sp := range salePayments {
		totalsMap[sp.SaleId] = sp.TotalPaidAmount
	}

	// add in flatData
	for i := range flatData {
		if flatData[i].SaleDetail != nil {
			flatData[i].SaleDetail.Paid = totalsMap[flatData[i].SaleDetail.Id]
			//flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice - totalsMap[flatData[i].SaleDetail.Id]
			flatData[i].SaleDetail.Remaining = flatData[i].SaleDetail.TotalPrice.Sub(totalsMap[flatData[i].SaleDetail.Id])
		}
	}

	return common.CreatePaginatedResponse(&flatData), nil
}

func (s *flatService) getSocietyFlatByName(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	name := r.URL.Query().Get("name")
	societyRera := chi.URLParam(r, "society")
	cursor := r.URL.Query().Get("cursor")

	flat := hGetSocietyFlatByName{}
	res, err := flat.execute(s.db, orgId, societyRera, name, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
