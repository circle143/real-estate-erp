package reports

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
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

type header struct {
	Heading string
	Items   []header
}

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

func (p paymentPlanInfo) getItems() []header {
	items := make([]header, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, header{
			Heading: item.Description,
			Items: []header{
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
	baseHeaders := []header{
		{
			Heading: "Flat details",
			Items: []header{
				{Heading: "Flat"},
				{Heading: "Floor"},
				{Heading: "Facing"},
				{Heading: "Saleable Area"},
				{Heading: "Unit Type"},
				{Heading: "Tower"},
			},
		},
		{
			Heading: "Sale Details",
			Items: []header{
				{Heading: "ID"},
				{Heading: "Total Price"},
				{Heading: "Total Payable Amount"},
				{Heading: "Paid Amount"},
				{Heading: "Pending Amount"},
			},
		},
		{
			Heading: "Broker Details",
			Items: []header{
				{Heading: "Name"},
				{Heading: "Aadhar"},
				{Heading: "PAN"},
			},
		},
		{
			Heading: "Payment Plan",
			Items: []header{
				{Heading: "Name"},
				{Heading: "Ratio"},
			},
		},
		{
			Heading: "Customer Details",
			Items: []header{
				{Heading: "Name"},
				{Heading: "Gender"},
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
			Heading: "Company Customer Details",
			Items: []header{
				{Heading: "Name"},
				{Heading: "Company PAN"},
				{Heading: "GST"},
				{Heading: "Aadhar"},
				{Heading: "PAN"},
			},
		},
	}

	// get unique sale price breakdown values
	salePriceBreakdown := make(map[string]bool)
	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil {
			for _, priceBreakdownItem := range flat.SaleDetail.PriceBreakdown {
				salePriceBreakdown[priceBreakdownItem.Summary] = true
			}
		}
	}
	salePriceBreakDownSlice := make([]header, 0, len(salePriceBreakdown))
	for breakdownItem := range salePriceBreakdown {
		salePriceBreakDownSlice = append(salePriceBreakDownSlice, header{
			Heading: breakdownItem,
		})
	}

	// add to baseHeaders
	baseHeaders = append(baseHeaders, header{
		Heading: "Price Breakdown",
		Items:   salePriceBreakDownSlice,
	})

	// get unique payment plans
	paymentPlanDetails := make(map[uuid.UUID]paymentPlanInfo)
	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil {
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

	// add to baseHeaders
	for _, item := range paymentPlanDetails {
		baseHeaders = append(baseHeaders, header{
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
		installmentItems := make([]header, 0, installmentCount)

		for i := 1; i <= installmentCount; i++ {
			installmentItems = append(installmentItems, header{
				Heading: strconv.Itoa(i),
				Items: []header{
					{Heading: "Date"},
					{Heading: "Amount"},
					{Heading: "Type"},
					{Heading: "Status"},
				},
			})
		}

		baseHeaders = append(baseHeaders, header{
			Heading: "Installment",
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

	_, err = renderHeaders(file, sheet, baseHeaders, 1, &colIndex, maxDepth, style)
	if err != nil {
		return err
	}

	for i := 1; i < colIndex; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		maxWidth := getMaxColumnWidth(file, sheet, colName, maxDepth)
		file.SetColWidth(sheet, colName, colName, maxWidth)
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
	return float64(maxLen) * 1.2
}

// Recursive function to render headers
func renderHeaders(f *excelize.File, sheet string, headers []header, row int, colIndex *int, maxDepth int, style int) (int, error) {
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
func getMaxDepth(headers []header, depth int) int {
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
	// sheetName := tower.Name
	// _, err := file.NewSheet(sheetName)
	// if err != nil {
	// 	return err
	// }
	//
	// // --- Step 1: Collect unique price breakdown summaries ---
	// summarySet := make(map[string]struct{})
	// for _, flat := range tower.Flats {
	// 	if flat.SaleDetail != nil {
	// 		for _, bd := range flat.SaleDetail.PriceBreakdown {
	// 			summarySet[bd.Summary] = struct{}{}
	// 		}
	// 	}
	// }
	// var priceBreakdownSummaries []string
	// for summary := range summarySet {
	// 	priceBreakdownSummaries = append(priceBreakdownSummaries, summary)
	// }
	// sort.Strings(priceBreakdownSummaries)
	//
	// // --- Step 2: Base headers ---
	// baseHeaders := []string{
	// 	"Tower_Name", "Flat_Name", "FloorNumber", "Facing", "SaleableArea", "UnitType",
	// 	"SaleNumber", "TotalPrice", "Paid", "Pending",
	// 	"BrokerName", "BrokerAadhar", "BrokerPan",
	// 	"PaymentPlanGroup", "PaymentPlanRatio",
	// 	"CustomerName", "CustomerEmail", "CustomerPhone", "CustomerPan", "CustomerAadhar",
	// 	"CompanyName", "CompanyPan", "CompanyGst",
	// }
	//
	// // --- Step 3: Write base headers merged over 3 rows ---
	// for i, h := range baseHeaders {
	// 	cell, _ := excelize.CoordinatesToCellName(i+1, 1)
	// 	if err := file.SetCellValue(sheetName, cell, h); err != nil {
	// 		return err
	// 	}
	// 	startCell, _ := excelize.CoordinatesToCellName(i+1, 1)
	// 	endCell, _ := excelize.CoordinatesToCellName(i+1, 3)
	// 	if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
	// 		return err
	// 	}
	// }
	//
	// // --- Step 4: PriceBreakdown headers ---
	// curCol := len(baseHeaders) + 1
	// if len(priceBreakdownSummaries) > 0 {
	// 	startCol := curCol
	// 	endCol := startCol + len(priceBreakdownSummaries) - 1
	//
	// 	// Row 1 merged "PriceBreakdown"
	// 	startCell, _ := excelize.CoordinatesToCellName(startCol, 1)
	// 	endCell, _ := excelize.CoordinatesToCellName(endCol, 1)
	// 	if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
	// 		return err
	// 	}
	// 	if err := file.SetCellValue(sheetName, startCell, "PriceBreakdown"); err != nil {
	// 		return err
	// 	}
	// 	style, _ := file.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center"}})
	// 	_ = file.SetCellStyle(sheetName, startCell, endCell, style)
	//
	// 	// Each summary spans row 2 → row 3
	// 	for j, summary := range priceBreakdownSummaries {
	// 		cell, _ := excelize.CoordinatesToCellName(startCol+j, 2)
	// 		if err := file.SetCellValue(sheetName, cell, summary); err != nil {
	// 			return err
	// 		}
	// 		startCell, _ := excelize.CoordinatesToCellName(startCol+j, 2)
	// 		endCell, _ := excelize.CoordinatesToCellName(startCol+j, 3)
	// 		if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
	// 			return err
	// 		}
	// 	}
	// 	curCol = endCol + 1
	// }
	//
	// // --- Step 5: PaymentPlanRatio headers ---
	// type ratioColumn struct {
	// 	Group string
	// 	Ratio string
	// 	Item  string
	// 	ID    uuid.UUID
	// 	Scope custom.PaymentPlanItemScope
	// }
	// var allRatios []ratioColumn
	// for _, flat := range tower.Flats {
	// 	if flat.SaleDetail != nil && flat.SaleDetail.PaymentPlanRatio != nil {
	// 		group := flat.SaleDetail.PaymentPlanRatio.PaymentPlanGroup
	// 		if group == nil {
	// 			continue
	// 		}
	// 		for _, item := range flat.SaleDetail.PaymentPlanRatio.Ratios {
	// 			allRatios = append(allRatios, ratioColumn{
	// 				Group: group.Name,
	// 				Ratio: flat.SaleDetail.PaymentPlanRatio.Ratio,
	// 				Item:  item.Description,
	// 				ID:    item.Id,
	// 				Scope: item.Scope,
	// 			})
	// 		}
	// 	}
	// }
	// // Deduplicate
	// seen := make(map[uuid.UUID]struct{})
	// var uniqueRatios []ratioColumn
	// for _, rc := range allRatios {
	// 	if _, ok := seen[rc.ID]; !ok {
	// 		seen[rc.ID] = struct{}{}
	// 		uniqueRatios = append(uniqueRatios, rc)
	// 	}
	// }
	//
	// if len(uniqueRatios) > 0 {
	// 	// Group by Group+Ratio
	// 	grouped := make(map[string][]ratioColumn)
	// 	var order []string
	// 	for _, rc := range uniqueRatios {
	// 		key := rc.Group + " - " + rc.Ratio
	// 		if _, ok := grouped[key]; !ok {
	// 			order = append(order, key)
	// 		}
	// 		grouped[key] = append(grouped[key], rc)
	// 	}
	//
	// 	for _, grKey := range order {
	// 		items := grouped[grKey]
	// 		startCol := curCol
	// 		endCol := startCol + len(items)*4 - 1
	//
	// 		// Row 1: group header
	// 		startCell, _ := excelize.CoordinatesToCellName(startCol, 1)
	// 		endCell, _ := excelize.CoordinatesToCellName(endCol, 1)
	// 		if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
	// 			return err
	// 		}
	// 		if err := file.SetCellValue(sheetName, startCell, grKey); err != nil {
	// 			return err
	// 		}
	// 		style, _ := file.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center"}})
	// 		_ = file.SetCellStyle(sheetName, startCell, endCell, style)
	//
	// 		for j, rc := range items {
	// 			itemStart := startCol + j*4
	// 			itemEnd := itemStart + 3
	//
	// 			// Row 2: item header
	// 			startCell, _ := excelize.CoordinatesToCellName(itemStart, 2)
	// 			endCell, _ := excelize.CoordinatesToCellName(itemEnd, 2)
	// 			if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
	// 				return err
	// 			}
	// 			if err := file.SetCellValue(sheetName, startCell, rc.Item); err != nil {
	// 				return err
	// 			}
	//
	// 			// Row 3: subheaders
	// 			for k, sub := range []string{"CollectionDate", "Total", "Pending", "Paid"} {
	// 				cell, _ := excelize.CoordinatesToCellName(itemStart+k, 3)
	// 				if err := file.SetCellValue(sheetName, cell, sub); err != nil {
	// 					return err
	// 				}
	// 			}
	// 		}
	// 		curCol = endCol + 1
	// 	}
	// }
	//
	// // --- Step 5.1: Pre-build status maps ---
	// towerStatusMap := make(map[uuid.UUID]time.Time)
	// for _, ts := range tower.ActivePaymentPlanRatioItems {
	// 	if ts.PaymentPlanRatioItem != nil {
	// 		towerStatusMap[ts.PaymentPlanRatioItem.Id] = ts.CreatedAt
	// 	}
	// }
	// flatStatusMap := make(map[string]time.Time)
	// for _, f := range tower.Flats {
	// 	for _, fs := range f.ActivePaymentPlanRatioItems {
	// 		if fs.PaymentPlanRatioItem != nil {
	// 			key := f.Id.String() + "-" + fs.PaymentPlanRatioItem.Id.String()
	// 			flatStatusMap[key] = fs.CreatedAt
	// 		}
	// 	}
	// }
	//
	// // --- Step 6: Data rows ---
	// rowIdx := 4
	// for _, flat := range tower.Flats {
	// 	baseFlat := []any{
	// 		sheetName,
	// 		flat.Name,
	// 		flat.FloorNumber,
	// 		flat.Facing,
	// 		flat.SaleableArea.String(),
	// 		flat.UnitType,
	// 	}
	//
	// 	if flat.SaleDetail != nil {
	// 		sale := flat.SaleDetail
	// 		saleVals := []any{
	// 			sale.SaleNumber,
	// 			sale.TotalPrice.String(),
	// 			sale.PaidAmount().String(),
	// 			sale.Pending().String(),
	// 		}
	//
	// 		var brokerVals []any
	// 		if sale.Broker != nil {
	// 			brokerVals = []any{sale.Broker.Name, sale.Broker.AadharNumber, sale.Broker.PanNumber}
	// 		} else {
	// 			brokerVals = []any{"-", "-", "-"}
	// 		}
	//
	// 		// payment plan group + ratio
	// 		var planVals []any
	// 		if sale.PaymentPlanRatio != nil && sale.PaymentPlanRatio.PaymentPlanGroup != nil {
	// 			planVals = []any{
	// 				sale.PaymentPlanRatio.PaymentPlanGroup.Name,
	// 				sale.PaymentPlanRatio.Ratio,
	// 			}
	// 		} else {
	// 			planVals = []any{"-", "-"}
	// 		}
	//
	// 		// Customers OR Company
	// 		if len(sale.Customers) > 0 {
	// 			for _, cust := range sale.Customers {
	// 				custVals := []any{
	// 					fmt.Sprintf("%s %s %s", cust.FirstName, cust.MiddleName, cust.LastName),
	// 					cust.Email,
	// 					cust.PhoneNumber,
	// 					cust.PanNumber,
	// 					cust.AadharNumber,
	// 				}
	// 				companyVals := []any{"-", "-", "-"}
	// 				rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
	// 				rowVals = append(rowVals, companyVals...)
	//
	// 				// Price breakdown values
	// 				bdMap := make(map[string]string)
	// 				for _, bd := range sale.PriceBreakdown {
	// 					bdMap[bd.Summary] = bd.Total.String()
	// 				}
	// 				for _, summary := range priceBreakdownSummaries {
	// 					if val, ok := bdMap[summary]; ok {
	// 						rowVals = append(rowVals, val)
	// 					} else {
	// 						rowVals = append(rowVals, "-")
	// 					}
	// 				}
	//
	// 				// Payment plan ratio items with CollectionDate
	// 				for _, rc := range uniqueRatios {
	// 					var collectionDate string
	// 					switch rc.Scope {
	// 					case custom.SCOPE_SALE:
	// 						collectionDate = sale.CreatedAt.Format("2006-01-02")
	// 					case custom.SCOPE_TOWER:
	// 						if dt, ok := towerStatusMap[rc.ID]; ok {
	// 							collectionDate = dt.Format("2006-01-02")
	// 						}
	// 					case custom.SCOPE_FLAT:
	// 						key := flat.Id.String() + "-" + rc.ID.String()
	// 						if dt, ok := flatStatusMap[key]; ok {
	// 							collectionDate = dt.Format("2006-01-02")
	// 						}
	// 					}
	// 					if collectionDate == "" {
	// 						collectionDate = "-"
	// 					}
	// 					rowVals = append(rowVals, collectionDate, "-", "-", "-")
	// 				}
	//
	// 				for colIdx, val := range rowVals {
	// 					cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
	// 					if err := file.SetCellValue(sheetName, cell, val); err != nil {
	// 						return err
	// 					}
	// 				}
	// 				rowIdx++
	// 			}
	// 		} else if sale.CompanyCustomer != nil {
	// 			custVals := []any{"-", "-", "-", "-", "-"}
	// 			companyVals := []any{
	// 				sale.CompanyCustomer.Name,
	// 				sale.CompanyCustomer.CompanyPan,
	// 				sale.CompanyCustomer.CompanyGst,
	// 			}
	// 			rowVals := append(append(append(baseFlat, saleVals...), brokerVals...), custVals...)
	// 			rowVals = append(rowVals, companyVals...)
	//
	// 			// Price breakdown
	// 			bdMap := make(map[string]string)
	// 			for _, bd := range sale.PriceBreakdown {
	// 				bdMap[bd.Summary] = bd.Total.String()
	// 			}
	// 			for _, summary := range priceBreakdownSummaries {
	// 				if val, ok := bdMap[summary]; ok {
	// 					rowVals = append(rowVals, val)
	// 				} else {
	// 					rowVals = append(rowVals, "-")
	// 				}
	// 			}
	//
	// 			// Payment plan ratio items with CollectionDate
	// 			for _, rc := range uniqueRatios {
	// 				var collectionDate string
	// 				switch rc.Scope {
	// 				case custom.SCOPE_SALE:
	// 					collectionDate = sale.CreatedAt.Format("2006-01-02")
	// 				case custom.SCOPE_TOWER:
	// 					if dt, ok := towerStatusMap[rc.ID]; ok {
	// 						collectionDate = dt.Format("2006-01-02")
	// 					}
	// 				case custom.SCOPE_FLAT:
	// 					key := flat.Id.String() + "-" + rc.ID.String()
	// 					if dt, ok := flatStatusMap[key]; ok {
	// 						collectionDate = dt.Format("2006-01-02")
	// 					}
	// 				}
	// 				if collectionDate == "" {
	// 					collectionDate = "-"
	// 				}
	// 				rowVals = append(rowVals, collectionDate, "-", "-", "-")
	// 			}
	//
	// 			for colIdx, val := range rowVals {
	// 				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
	// 				if err := file.SetCellValue(sheetName, cell, val); err != nil {
	// 					return err
	// 				}
	// 			}
	// 			rowIdx++
	// 		}
	// 	} else {
	// 		// Flat without sale
	// 		emptySale := []any{"-", "-", "-", "-"}
	// 		emptyBroker := []any{"-", "-", "-"}
	// 		emptyCust := []any{"-", "-", "-", "-", "-"}
	// 		emptyCompany := []any{"-", "-", "-"}
	// 		rowVals := append(append(append(baseFlat, emptySale...), emptyBroker...), emptyCust...)
	// 		rowVals = append(rowVals, emptyCompany...)
	// 		for range priceBreakdownSummaries {
	// 			rowVals = append(rowVals, "-")
	// 		}
	// 		for range uniqueRatios {
	// 			rowVals = append(rowVals, "-", "-", "-", "-")
	// 		}
	// 		for colIdx, val := range rowVals {
	// 			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
	// 			if err := file.SetCellValue(sheetName, cell, val); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		rowIdx++
	// 	}
	// }
	//
	// return nil
}

func generateMasterReport(db *gorm.DB, orgId, society string) (*bytes.Buffer, error) {
	var towerData []models.Tower
	err := db.
		Where("org_id = ? AND society_id = ?", orgId, society).
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

	var reportFile *excelize.File = nil
	if os.Getenv("ENV") == "development" {
		reportFile = excelize.NewFile()
	} else {
		reportFile = excelize.NewFile(
			excelize.Options{
				Password: society,
			},
		)
	}

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

	report, err := generateMasterReport(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	// Set headers so browser/download tools recognize it as Excel
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_master_report_%d.xlsx", societyRera, time.Now().Unix()))
	w.Header().Set("Content-Length", fmt.Sprint(report.Len()))

	// Write to response
	if _, err := w.Write(report.Bytes()); err != nil {
		payload.HandleError(w, err)
		return
	}
}
