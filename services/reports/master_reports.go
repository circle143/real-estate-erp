package reports

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sort"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func newMasterReportSheet(file *excelize.File, tower models.Tower) error {
	sheetName := tower.Name
	_, err := file.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// --- Step 1: Collect unique price breakdown summaries ---
	summarySet := make(map[string]struct{})
	for _, flat := range tower.Flats {
		if flat.SaleDetail != nil {
			for _, bd := range flat.SaleDetail.PriceBreakdown {
				summarySet[bd.Summary] = struct{}{}
			}
		}
	}
	var priceBreakdownSummaries []string
	for summary := range summarySet {
		priceBreakdownSummaries = append(priceBreakdownSummaries, summary)
	}
	sort.Strings(priceBreakdownSummaries) // consistent column order

	// --- Step 2: Base headers ---
	baseHeaders := []string{
		"Tower_Name", "Flat_Name", "FloorNumber", "Facing", "SaleableArea", "UnitType",
		"SaleNumber", "TotalPrice", "Paid", "Pending",
		"BrokerName", "BrokerAadhar", "BrokerPan",
		"PaymentPlanGroup", "PaymentPlanRatio",
		"CustomerName", "CustomerEmail", "CustomerPhone", "CustomerPan", "CustomerAadhar",
		"CompanyName", "CompanyPan", "CompanyGst",
	}

	// --- Step 3: Write base headers in row 1 and merge down to row 2 ---
	for i, h := range baseHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := file.SetCellValue(sheetName, cell, h); err != nil {
			return err
		}
		startCell, _ := excelize.CoordinatesToCellName(i+1, 1)
		endCell, _ := excelize.CoordinatesToCellName(i+1, 2)
		if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
			return err
		}
	}

	// --- Step 4: Add PriceBreakdown merged header + summaries ---
	if len(priceBreakdownSummaries) > 0 {
		startCol := len(baseHeaders) + 1
		endCol := startCol + len(priceBreakdownSummaries) - 1

		// Merge row 1 for "PriceBreakdown"
		startCell, _ := excelize.CoordinatesToCellName(startCol, 1)
		endCell, _ := excelize.CoordinatesToCellName(endCol, 1)
		if err := file.MergeCell(sheetName, startCell, endCell); err != nil {
			return err
		}
		if err := file.SetCellValue(sheetName, startCell, "PriceBreakdown"); err != nil {
			return err
		}

		// Create a centered bold style
		styleID, err := file.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
			Font: &excelize.Font{
				Bold: true,
			},
		})
		if err != nil {
			return err
		}

		// Apply style to merged range
		if err := file.SetCellStyle(sheetName, startCell, endCell, styleID); err != nil {
			return err
		}

		// Row 2: each summary name
		for j, summary := range priceBreakdownSummaries {
			cell, _ := excelize.CoordinatesToCellName(startCol+j, 2)
			if err := file.SetCellValue(sheetName, cell, summary); err != nil {
				return err
			}
		}
	}

	// --- Step 5: Data rows (start from row 3 now) ---
	rowIdx := 3
	for _, flat := range tower.Flats {
		baseFlat := []any{
			sheetName,
			flat.Name,
			flat.FloorNumber,
			flat.Facing,
			flat.SaleableArea.String(),
			flat.UnitType,
		}

		if flat.SaleDetail != nil {
			sale := flat.SaleDetail

			saleVals := []any{
				sale.SaleNumber,
				sale.TotalPrice.String(),
				sale.PaidAmount().String(),
				sale.Pending().String(),
			}

			var brokerVals []any
			if sale.Broker != nil {
				brokerVals = []any{
					sale.Broker.Name,
					sale.Broker.AadharNumber,
					sale.Broker.PanNumber,
				}
			} else {
				brokerVals = []any{"-", "-", "-"}
			}

			var planVals []any
			if sale.PaymentPlanRatio != nil && sale.PaymentPlanRatio.PaymentPlanGroup != nil {
				planVals = []any{
					sale.PaymentPlanRatio.PaymentPlanGroup.Name,
					sale.PaymentPlanRatio.Ratio,
				}
			} else {
				planVals = []any{"-", "-"}
			}

			// map summary → total
			breakdownMap := make(map[string]string)
			for _, bd := range sale.PriceBreakdown {
				breakdownMap[bd.Summary] = bd.Total.String()
			}

			// Customers
			if len(sale.Customers) > 0 {
				for _, cust := range sale.Customers {
					custVals := []any{
						fmt.Sprintf("%s %s %s", cust.FirstName, cust.MiddleName, cust.LastName),
						cust.Email,
						cust.PhoneNumber,
						cust.PanNumber,
						cust.AadharNumber,
					}
					companyVals := []any{"-", "-", "-"}
					rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
					rowVals = append(rowVals, companyVals...)

					// add breakdown totals
					for _, summary := range priceBreakdownSummaries {
						if val, ok := breakdownMap[summary]; ok {
							rowVals = append(rowVals, val)
						} else {
							rowVals = append(rowVals, "-")
						}
					}

					for colIdx, val := range rowVals {
						cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
						if err := file.SetCellValue(sheetName, cell, val); err != nil {
							return err
						}
					}
					rowIdx++
				}
			} else if sale.CompanyCustomer != nil {
				custVals := []any{"-", "-", "-", "-", "-"}
				companyVals := []any{
					sale.CompanyCustomer.Name,
					sale.CompanyCustomer.CompanyPan,
					sale.CompanyCustomer.CompanyGst,
				}
				rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
				rowVals = append(rowVals, companyVals...)
				for _, summary := range priceBreakdownSummaries {
					if val, ok := breakdownMap[summary]; ok {
						rowVals = append(rowVals, val)
					} else {
						rowVals = append(rowVals, "-")
					}
				}
				for colIdx, val := range rowVals {
					cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
					if err := file.SetCellValue(sheetName, cell, val); err != nil {
						return err
					}
				}
				rowIdx++
			}
		} else {
			// flat without sale
			emptySale := []any{"-", "-"}
			emptyBroker := []any{"-", "-", "-"}
			emptyPlan := []any{"-", "-"}
			emptyCust := []any{"-", "-", "-", "-", "-"}
			emptyCompany := []any{"-", "-", "-"}
			rowVals := append(append(append(append(baseFlat, emptySale...), emptyBroker...), emptyPlan...), emptyCust...)
			rowVals = append(rowVals, emptyCompany...)
			for range priceBreakdownSummaries {
				rowVals = append(rowVals, "-")
			}
			for colIdx, val := range rowVals {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
				if err := file.SetCellValue(sheetName, cell, val); err != nil {
					return err
				}
			}
			rowIdx++
		}
	}

	return nil
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_master_report.xlsx", societyRera))
	w.Header().Set("Content-Length", fmt.Sprint(report.Len()))

	// Write to response
	if _, err := w.Write(report.Bytes()); err != nil {
		payload.HandleError(w, err)
		return
	}
}
