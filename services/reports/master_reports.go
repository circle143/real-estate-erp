package reports

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type paymentPlanItemInfo struct {
	ID          uuid.UUID
	Description string
}

type paymentPlanInfo struct {
	ID    uuid.UUID
	Name  string
	Ratio string
	Items []paymentPlanItemInfo
}

func (p paymentPlanInfo) getHeading() string {
	return fmt.Sprintf("%s (%s)", p.Name, p.Ratio)
}

func (p paymentPlanInfo) getItems() []models.Header {
	items := make([]models.Header, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, models.Header{
			ID:      &item.ID,
			Heading: item.Description,
			Items: []models.Header{
				{Heading: "Collection Date"},
				{Heading: "Total Amount"},
				{Heading: "Paid"},
				{Heading: "Pending"},
			},
		})
	}

	return items
}

func newMasterReportSheetManual(file *excelize.File, tower models.Tower) error {
	sheet := tower.Name
	_, err := file.NewSheet(sheet)
	if err != nil {
		return err
	}

	// base headers
	baseHeaders := []models.Header{
		{
			Heading: models.HeadingFlat,
			Items: []models.Header{
				{Heading: "Tower"},
				{Heading: "Floor"},
				{Heading: "Flat"},
				{Heading: "Facing"},
				{Heading: "Unit Type"},
				{Heading: "Saleable Area"},
			},
		},
		{
			Heading: models.HeadingPaymentPlan,
			Items: []models.Header{
				{Heading: "Name"},
				// {Heading: "Ratio"},
			},
		},
		{
			Heading: models.HeadingCustomer,
			Items: []models.Header{
				{Heading: "Name"},
				// {Heading: "Gender"},
				{Heading: "Email"},
				{Heading: "Phone Number"},
				{Heading: "Nationality"},
				{Heading: "Aadhar"},
				{Heading: "PAN"},
				{Heading: "Passport Number"},
				{Heading: "Profession"},
				{Heading: "Company Name"},
			},
		},
		{
			Heading: models.HeadingCompanyCustomer,
			Items: []models.Header{
				{Heading: "Name"},
				{Heading: "Company PAN"},
				{Heading: "GST"},
				{Heading: "Aadhar"},
				{Heading: "PAN"},
			},
		},
	}

	// get unique sale price breakdown values while preserving order
	salePriceBreakdownSet := make(map[string]bool)
	salePriceBreakDownSlice := make([]models.Header, 0)
	var ifmsHeader *models.Header

	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil {
			for _, priceBreakdownItem := range flat.SaleDetail.PriceBreakdown {
				summary := priceBreakdownItem.Summary
				if !salePriceBreakdownSet[summary] {
					salePriceBreakdownSet[summary] = true

					if summary == "Intrest Free Maintenance Security (IFMS)" {
						// defer adding IFMS until the end
						ifmsHeader = &models.Header{Heading: summary}
					} else {
						salePriceBreakDownSlice = append(salePriceBreakDownSlice, models.Header{
							Heading: summary,
						})
					}
				}
			}
		}
	}

	// finally, add IFMS at the end if it exists
	if ifmsHeader != nil {
		salePriceBreakDownSlice = append(salePriceBreakDownSlice, *ifmsHeader)
	}

	// add to basemodels.Headers
	baseHeaders = append(baseHeaders, models.Header{
		Heading: models.HeadingPricebreakdown,
		Items:   salePriceBreakDownSlice,
	})

	// add broker details
	baseHeaders = append(baseHeaders,
		[]models.Header{
			{
				Heading: models.HeadingSale,
				Items: []models.Header{
					// {Heading: "ID"},
					{Heading: "Total Price"},
					{Heading: "Total Payable Amount"},
					{Heading: "Total Paid Amount"},
					{Heading: "Paid Amount"},
					{Heading: "Pending Amount"},
					{Heading: "CGST"},
					{Heading: "SGST"},
					{Heading: "Service Tax"},
					{Heading: "Swathch Bharat Cess"},
					{Heading: "Krishi Kalyan Cess"},
				},
			},
			{
				Heading: models.HeadingBroker,
				Items: []models.Header{
					{Heading: "Name"},
					{Heading: "Aadhar"},
					{Heading: "PAN"},
				},
			},
		}...,
	)

	// get unique payment plans
	paymentPlanDetails := make(map[uuid.UUID]paymentPlanInfo)
	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil && flat.SaleDetail.PaymentPlanRatio != nil {
			ratioKey := flat.SaleDetail.PaymentPlanRatioId
			ratioItems := make([]paymentPlanItemInfo, 0, len(flat.SaleDetail.PaymentPlanRatio.Ratios))

			for _, ratioItem := range flat.SaleDetail.PaymentPlanRatio.Ratios {
				ratioItems = append(ratioItems, paymentPlanItemInfo{
					ID:          ratioItem.Id,
					Description: ratioItem.Description,
				})
			}

			paymentPlanDetails[ratioKey] = paymentPlanInfo{
				ID:    ratioKey,
				Name:  flat.SaleDetail.PaymentPlanRatio.PaymentPlanGroup.Name,
				Ratio: flat.SaleDetail.PaymentPlanRatio.Ratio,
				Items: ratioItems,
			}
		}
	}

	// add to basemodels.Headers
	for _, item := range paymentPlanDetails {
		baseHeaders = append(baseHeaders, models.Header{
			ID:      &item.ID,
			Heading: item.getHeading(),
			Items:   item.getItems(),
		})
	}

	// get max valid installment number
	installmentCount := 0
	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil {
			installmentCount = max(installmentCount, flat.SaleDetail.GetValidReceiptsCount())
		}
	}

	if installmentCount > 0 {
		// add installment header
		installmentItems := make([]models.Header, 0, installmentCount)

		for i := 1; i <= installmentCount; i++ {
			installmentItems = append(installmentItems, models.Header{
				Heading: strconv.Itoa(i),
				Items: []models.Header{
					{Heading: "Number"},
					{Heading: "Date"},
					{Heading: "Amount"},
					{Heading: "Type"},
					{Heading: "CGST"},
					{Heading: "SGST"},
					{Heading: "Service Tax"},
					{Heading: "Swathch Bharat Cess"},
					{Heading: "Krishi Kalyan Cess"},
					{Heading: "Status"},
					{Heading: "Cleared At"},
				},
			})
		}

		baseHeaders = append(baseHeaders, models.Header{
			Heading: models.HeadingInstallment,
			Items:   installmentItems,
		})
	}

	// Create a style for centered, bold headers
	style, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}

	maxDepth := getMaxDepth(baseHeaders, 1)
	colIndex := 1

	headers := []models.Header{
		{
			Heading: "Member ID",
		},
	}

	headers = append(headers, baseHeaders...)

	_, err = renderHeaders(file, sheet, headers, 1, &colIndex, maxDepth, style)
	if err != nil {
		return err
	}

	for i := 1; i < colIndex; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		maxWidth := getMaxColumnWidth(file, sheet, colName, maxDepth)
		file.SetColWidth(sheet, colName, colName, maxWidth)
	}

	startRow := getMaxDepth(baseHeaders, 1) + 1

	for i, flat := range tower.Flats {
		rowNum := startRow + i
		values := flat.GetRowData(baseHeaders, sheet, models.SafePrint{
			ShouldPrint: i == 0 && sheet == "A",
		}, tower.ActivePaymentPlanRatioItems)
		for colIdx, val := range values {
			colName, _ := excelize.ColumnNumberToName(colIdx + 1)
			cell := fmt.Sprintf("%s%d", colName, rowNum)
			file.SetCellValue(sheet, cell, val)
		}
	}

	return nil
}

// Helper: find max string length in a column
func getMaxColumnWidth(f *excelize.File, sheet, col string, maxRow int) float64 {
	maxLen := 0
	for row := 1; row <= maxRow; row++ {
		cell := fmt.Sprintf("%s%d", col, row)
		val, _ := f.GetCellValue(sheet, cell)
		if len(val) > maxLen {
			maxLen = len(val)
		}
	}
	// Multiply by 1.2 for padding
	return float64(min(maxLen, 15)) * 1.2
}

// Recursive function to render headers
func renderHeaders(f *excelize.File, sheet string, headers []models.Header, row int, colIndex *int, maxDepth int, style int) (int, error) {
	for _, h := range headers {
		startCol := *colIndex
		if len(h.Items) > 0 {
			// Has children → render recursively
			_, err := renderHeaders(f, sheet, h.Items, row+1, colIndex, maxDepth, style)
			if err != nil {
				return 0, err
			}
		} else {
			*colIndex++ // leaf header occupies one column
		}
		endCol := *colIndex - 1

		// Merge horizontally if more than one column
		startColName, _ := excelize.ColumnNumberToName(startCol)
		endColName, _ := excelize.ColumnNumberToName(endCol)
		if endCol > startCol {
			if err := f.MergeCell(sheet, fmt.Sprintf("%s%d", startColName, row), fmt.Sprintf("%s%d", endColName, row)); err != nil {
				return 0, err
			}
		}

		// Merge vertically if leaf header but not at max depth
		if len(h.Items) == 0 && row < maxDepth {
			if err := f.MergeCell(sheet, fmt.Sprintf("%s%d", startColName, row), fmt.Sprintf("%s%d", startColName, maxDepth)); err != nil {
				return 0, err
			}
		}

		// Set header value and style
		cell := fmt.Sprintf("%s%d", startColName, row)
		f.SetCellValue(sheet, cell, h.Heading)
		f.SetCellStyle(sheet, cell, cell, style)
	}

	return *colIndex, nil
}

// Helper: calculate max depth of nested headers
func getMaxDepth(headers []models.Header, depth int) int {
	max := depth
	for _, h := range headers {
		if len(h.Items) > 0 {
			d := getMaxDepth(h.Items, depth+1)
			if d > max {
				max = d
			}
		}
	}
	return max
}

func newMasterReportSheet(file *excelize.File, tower models.Tower) error {
	return newMasterReportSheetManual(file, tower)
}

func generateMasterReport(db *gorm.DB, orgId, society, tower string) (*bytes.Buffer, error) {

	query := db.
		Where("org_id = ? AND society_id = ?", orgId, society)

	if tower != "" {
		query = query.Where("name = ?", tower)
	}

	var towerData []models.Tower
	err := query.
		Preload("ActivePaymentPlanRatioItems").
		Preload("Flats").
		Preload("Flats.ActivePaymentPlanRatioItems").
		Preload("Flats.SaleDetail").
		Preload("Flats.SaleDetail.PaymentPlanRatio").
		Preload("Flats.SaleDetail.PaymentPlanRatio.PaymentPlanGroup").
		Preload("Flats.SaleDetail.PaymentPlanRatio.Ratios").
		Preload("Flats.SaleDetail.Receipts").
		Preload("Flats.SaleDetail.Receipts.Cleared").
		Preload("Flats.SaleDetail.Broker").
		Preload("Flats.SaleDetail.Customers").
		Preload("Flats.SaleDetail.CompanyCustomer").
		Find(&towerData).Error
	if err != nil {
		return nil, err
	}

	if len(towerData) < 1 {
		return nil, &custom.RequestError{
			Status:  http.StatusNoContent,
			Message: "No tower found",
		}
	}

	reportFile := excelize.NewFile()

	for _, tower := range towerData {
		sheetErr := newMasterReportSheet(reportFile, tower)
		if sheetErr != nil {
			return nil, sheetErr
		}
	}

	// Delete the default "Sheet1"
	if err := reportFile.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	// Write Excel file to buffer
	var buf bytes.Buffer
	if err := reportFile.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

// generateMasterReport() generates a master report
// master report contains all the flats and its related sale details
func (s *reportService) generateMasterReport(w http.ResponseWriter, r *http.Request) {
	// get society rera and org id from request
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	tower := r.URL.Query().Get("tower")

	report, err := generateMasterReport(s.db, orgId, societyRera, tower)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	fileNameBase := societyRera
	if tower != "" {
		fileNameBase = fmt.Sprintf("tower_%s", tower)
	}

	// Set headers so browser/download tools recognize it as Excel
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%s_master_report_%d.xlsx", fileNameBase, time.Now().Unix()),
	)
	w.Header().Set("Content-Length", fmt.Sprint(report.Len()))

	// Write to response
	if _, err := w.Write(report.Bytes()); err != nil {
		payload.HandleError(w, err)
		return
	}
}
