package reports

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Indian number format without symbol: 1,00,00,000
const IndianNumberFormat = `[>=10000000]##\,##\,##\,##0;[>=100000]##\,##\,##0;##,##0`

// Alternative format with decimals: 1,00,00,000.00
const IndianNumberFormatDecimal = `[>=10000000]##\,##\,##\,##0.00;[>=100000]##\,##\,##0.00;##,##0.00`

// List of monetary field keywords to identify money columns
var monetaryFieldKeywords = []string{
	"amount", "price", "total", "paid", "pending", "payable",
	"cgst", "sgst", "service tax", "swathch bharat cess", "krishi kalyan cess",
	"tax", "cess", "gst", "fee", "charges", "cost", "value", "payment",
	"collection", "receipt", "balance", "due", "outstanding", "security",
	"maintenance", "ifms", "₹", "rs", "inr",
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

func (p paymentPlanInfo) getItems() []models.Header {
	items := make([]models.Header, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, models.Header{
			ID:      &item.ID,
			Heading: item.Description,
			Items: []models.Header{
				{Heading: "Collection Date"},
				{Heading: "Total Amount", IsMonetary: true},
				{Heading: "Paid", IsMonetary: true},
				{Heading: "Pending", IsMonetary: true},
			},
		})
	}

	return items
}

// isMonetaryColumn checks if a column header represents a monetary field
func isMonetaryColumn(heading string) bool {
	headingLower := strings.ToLower(heading)
	for _, keyword := range monetaryFieldKeywords {
		if strings.Contains(headingLower, keyword) {
			return true
		}
	}
	return false
}

// getMonetaryColumnIndices returns a map of column indices that contain monetary data
func getMonetaryColumnIndices(headers []models.Header) map[int]bool {
	monetaryColumns := make(map[int]bool)
	colIndex := 1 // Start from 1 (Excel columns are 1-indexed)

	var traverse func(hs []models.Header)
	traverse = func(hs []models.Header) {
		for _, h := range hs {
			if len(h.Items) > 0 {
				traverse(h.Items)
			} else {
				// Leaf node - check if monetary
				if h.IsMonetary || isMonetaryColumn(h.Heading) {
					monetaryColumns[colIndex] = true
				}
				colIndex++
			}
		}
	}
	traverse(headers)

	return monetaryColumns
}

// getYellowMonetaryColumnIndices returns indices of monetary columns with yellow color
func getYellowMonetaryColumnIndices(headers []models.Header) map[int]bool {
	yellowColumns := make(map[int]bool)
	colIndex := 1

	var traverse func(hs []models.Header)
	traverse = func(hs []models.Header) {
		for _, h := range hs {
			if len(h.Items) > 0 {
				traverse(h.Items)
			} else {
				if (h.IsMonetary || isMonetaryColumn(h.Heading)) && strings.ToLower(h.Color) == "yellow" {
					yellowColumns[colIndex] = true
				}
				colIndex++
			}
		}
	}
	traverse(headers)

	return yellowColumns
}

// createNumberStyle creates an Excel style for Indian number format (no symbol)
func createNumberStyle(file *excelize.File) (int, error) {
	return file.NewStyle(&excelize.Style{
		NumFmt: 0,
		CustomNumFmt: func() *string {
			s := IndianNumberFormat
			return &s
		}(),
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
	})
}

// createNumberStyleWithColor creates number style with background color
func createNumberStyleWithColor(file *excelize.File, colorHex string) (int, error) {
	return file.NewStyle(&excelize.Style{
		CustomNumFmt: func() *string {
			s := IndianNumberFormat
			return &s
		}(),
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colorHex},
			Pattern: 1,
		},
	})
}

// parseToFloat converts a string value to float64
// Returns the float value and true if successful, 0 and false otherwise
func parseToFloat(val string) (float64, bool) {
	// Trim whitespace
	s := strings.TrimSpace(val)

	// Skip empty or placeholder values
	if s == "" || s == "-" || s == "N/A" || s == "NA" || s == "null" || s == "nil" {
		return 0, false
	}

	// Remove currency symbols and formatting
	s = strings.ReplaceAll(s, "₹", "")
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	// Try to parse as float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}

	return 0, false
}

func newMasterReportSheetManual(file *excelize.File, tower models.Tower) error {
	sheet := tower.Name
	_, err := file.NewSheet(sheet)
	if err != nil {
		return err
	}

	// base headers with IsMonetary flag for monetary columns
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
			},
		},
		{
			Heading: models.HeadingCustomer,
			Items: []models.Header{
				{Heading: "Name"},
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

					// All price breakdown items are monetary
					header := models.Header{
						Heading:    summary,
						IsMonetary: true,
					}

					if summary == "Intrest Free Maintenance Security (IFMS)" {
						ifmsHeader = &header
					} else {
						salePriceBreakDownSlice = append(salePriceBreakDownSlice, header)
					}
				}
			}
		}
	}

	if ifmsHeader != nil {
		salePriceBreakDownSlice = append(salePriceBreakDownSlice, *ifmsHeader)
	}

	baseHeaders = append(baseHeaders, models.Header{
		Heading: models.HeadingPricebreakdown,
		Items:   salePriceBreakDownSlice,
	})

	// add broker details - mark monetary columns
	baseHeaders = append(baseHeaders,
		[]models.Header{
			{
				Heading: models.HeadingSale,
				Items: []models.Header{
					{Heading: "Total Price", IsMonetary: true},
					{Heading: "Total Payable Amount", IsMonetary: true},
					{Heading: "Total Paid Amount", IsMonetary: true},
					{Heading: "Paid Amount", Color: "yellow", IsMonetary: true},
					{Heading: "CGST", Color: "yellow", IsMonetary: true},
					{Heading: "SGST", Color: "yellow", IsMonetary: true},
					{Heading: "Service Tax", Color: "yellow", IsMonetary: true},
					{Heading: "Swathch Bharat Cess", Color: "yellow", IsMonetary: true},
					{Heading: "Krishi Kalyan Cess", Color: "yellow", IsMonetary: true},
					{Heading: "Pending Amount", IsMonetary: true},
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

	// add payment plan headers (these have monetary sub-items defined in getItems())
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
		installmentItems := make([]models.Header, 0, installmentCount)

		for i := 1; i <= installmentCount; i++ {
			installmentItems = append(installmentItems, models.Header{
				Heading: strconv.Itoa(i),
				Items: []models.Header{
					{Heading: "Number"},
					{Heading: "Date"},
					{Heading: "Amount", IsMonetary: true},
					{Heading: "Type"},
					{Heading: "CGST", IsMonetary: true},
					{Heading: "SGST", IsMonetary: true},
					{Heading: "Service Tax", IsMonetary: true},
					{Heading: "Swathch Bharat Cess", IsMonetary: true},
					{Heading: "Krishi Kalyan Cess", IsMonetary: true},
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

	// Create base style for headers
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}

	// Create number style for monetary cells (Indian format without symbol)
	numberStyle, err := createNumberStyle(file)
	if err != nil {
		return err
	}

	// Create number style with yellow background
	numberStyleYellow, err := createNumberStyleWithColor(file, "#FFFF00")
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

	_, err = renderHeaders(file, sheet, headers, 1, &colIndex, maxDepth, headerStyle)
	if err != nil {
		return err
	}

	// Set column widths
	for i := 1; i < colIndex; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		maxWidth := getMaxColumnWidth(file, sheet, colName, maxDepth)
		file.SetColWidth(sheet, colName, colName, maxWidth)
	}

	// Get monetary column indices
	monetaryColumns := getMonetaryColumnIndices(headers)

	// Get columns that should have yellow background
	yellowMonetaryColumns := getYellowMonetaryColumnIndices(headers)

	startRow := maxDepth + 1

	for i, flat := range tower.Flats {
		rowNum := startRow + i
		values := flat.GetRowData(baseHeaders, sheet, models.SafePrint{
			ShouldPrint: i == 0 && sheet == "A",
		}, tower.ActivePaymentPlanRatioItems)

		for colIdx, val := range values {
			colNum := colIdx + 1
			colName, _ := excelize.ColumnNumberToName(colNum)
			cell := fmt.Sprintf("%s%d", colName, rowNum)

			// Check if this is a monetary column
			if monetaryColumns[colNum] {
				// Try to convert string value to float for proper Excel number formatting
				if numVal, ok := parseToFloat(val); ok {
					// Set as numeric value - this is KEY for Excel formatting to work
					file.SetCellFloat(sheet, cell, numVal, -1, 64)

					// Apply appropriate number style
					if yellowMonetaryColumns[colNum] {
						file.SetCellStyle(sheet, cell, cell, numberStyleYellow)
					} else {
						file.SetCellStyle(sheet, cell, cell, numberStyle)
					}
				} else {
					// If not a valid number (like "-"), set as string
					file.SetCellValue(sheet, cell, val)
				}
			} else {
				file.SetCellValue(sheet, cell, val)
			}
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
	// Multiply by 1.2 for padding, min 12 for currency columns
	width := float64(min(maxLen, 15)) * 1.2
	if width < 12 {
		width = 12 // Minimum width for currency display
	}
	return width
}

func renderHeaders(f *excelize.File, sheet string, headers []models.Header, row int, colIndex *int, maxDepth int, baseStyle int) (int, error) {
	styleCache := make(map[string]int)

	var applyHeader func([]models.Header, int, string) (int, error)
	applyHeader = func(hs []models.Header, currentRow int, parent string) (int, error) {
		for _, h := range hs {
			startCol := *colIndex
			if len(h.Items) > 0 {
				_, err := applyHeader(h.Items, currentRow+1, h.Heading)
				if err != nil {
					return 0, err
				}
			} else {
				*colIndex++
			}
			endCol := *colIndex - 1

			startColName, _ := excelize.ColumnNumberToName(startCol)
			endColName, _ := excelize.ColumnNumberToName(endCol)

			if endCol > startCol {
				if err := f.MergeCell(sheet, fmt.Sprintf("%s%d", startColName, currentRow),
					fmt.Sprintf("%s%d", endColName, currentRow)); err != nil {
					return 0, err
				}
			}

			if len(h.Items) == 0 && currentRow < maxDepth {
				if err := f.MergeCell(sheet,
					fmt.Sprintf("%s%d", startColName, currentRow),
					fmt.Sprintf("%s%d", startColName, maxDepth)); err != nil {
					return 0, err
				}
			}

			cell := fmt.Sprintf("%s%d", startColName, currentRow)
			f.SetCellValue(sheet, cell, h.Heading)

			styleID := baseStyle
			if h.Color != "" {
				colorName := strings.ToLower(h.Color)

				if cached, ok := styleCache[colorName]; ok {
					styleID = cached
				} else {
					colorHex := map[string]string{
						"yellow": "#FFFF00",
						"green":  "#00FF00",
						"red":    "#FF0000",
						"blue":   "#007BFF",
						"gray":   "#D3D3D3",
					}[colorName]
					if colorHex == "" {
						colorHex = h.Color
					}

					styleJSON, _ := f.GetStyle(baseStyle)
					newStyle := *styleJSON

					newStyle.Fill = excelize.Fill{
						Type:    "pattern",
						Color:   []string{colorHex},
						Pattern: 1,
					}

					newStyle.Border = []excelize.Border{
						{Type: "left", Color: "000000", Style: 1},
						{Type: "right", Color: "000000", Style: 1},
						{Type: "top", Color: "000000", Style: 1},
						{Type: "bottom", Color: "000000", Style: 1},
					}

					newStyleID, err := f.NewStyle(&newStyle)
					if err != nil {
						return 0, err
					}

					styleCache[colorName] = newStyleID
					styleID = newStyleID
				}
			}

			f.SetCellStyle(sheet, cell, cell, styleID)
		}
		return *colIndex, nil
	}

	return applyHeader(headers, row, "")
}

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

// newMasterReportSheetAllTowers creates a single sheet with all towers' data
func newMasterReportSheetAllTowers(file *excelize.File, towers []models.Tower) error {
	sheet := "Master Report"
	_, err := file.NewSheet(sheet)
	if err != nil {
		return err
	}

	// Combine all flats from all towers for header generation
	var allFlats []models.Flat
	for _, tower := range towers {
		allFlats = append(allFlats, tower.Flats...)
	}

	// base headers with IsMonetary flag for monetary columns
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
			},
		},
		{
			Heading: models.HeadingCustomer,
			Items: []models.Header{
				{Heading: "Name"},
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

	// get unique sale price breakdown values from ALL towers
	salePriceBreakdownSet := make(map[string]bool)
	salePriceBreakDownSlice := make([]models.Header, 0)
	var ifmsHeader *models.Header

	for _, flat := range allFlats {
		if flat.SaleDetail != nil {
			for _, priceBreakdownItem := range flat.SaleDetail.PriceBreakdown {
				summary := priceBreakdownItem.Summary
				if !salePriceBreakdownSet[summary] {
					salePriceBreakdownSet[summary] = true

					header := models.Header{
						Heading:    summary,
						IsMonetary: true,
					}

					if summary == "Intrest Free Maintenance Security (IFMS)" {
						ifmsHeader = &header
					} else {
						salePriceBreakDownSlice = append(salePriceBreakDownSlice, header)
					}
				}
			}
		}
	}

	if ifmsHeader != nil {
		salePriceBreakDownSlice = append(salePriceBreakDownSlice, *ifmsHeader)
	}

	baseHeaders = append(baseHeaders, models.Header{
		Heading: models.HeadingPricebreakdown,
		Items:   salePriceBreakDownSlice,
	})

	// add sale and broker details
	baseHeaders = append(baseHeaders,
		[]models.Header{
			{
				Heading: models.HeadingSale,
				Items: []models.Header{
					{Heading: "Total Price", IsMonetary: true},
					{Heading: "Total Payable Amount", IsMonetary: true},
					{Heading: "Total Paid Amount", IsMonetary: true},
					{Heading: "Paid Amount", Color: "yellow", IsMonetary: true},
					{Heading: "CGST", Color: "yellow", IsMonetary: true},
					{Heading: "SGST", Color: "yellow", IsMonetary: true},
					{Heading: "Service Tax", Color: "yellow", IsMonetary: true},
					{Heading: "Swathch Bharat Cess", Color: "yellow", IsMonetary: true},
					{Heading: "Krishi Kalyan Cess", Color: "yellow", IsMonetary: true},
					{Heading: "Pending Amount", IsMonetary: true},
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

	// get unique payment plans from ALL towers
	paymentPlanDetails := make(map[uuid.UUID]paymentPlanInfo)
	for _, flat := range allFlats {
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

	for _, item := range paymentPlanDetails {
		baseHeaders = append(baseHeaders, models.Header{
			ID:      &item.ID,
			Heading: item.getHeading(),
			Items:   item.getItems(),
		})
	}

	// get max valid installment number from ALL towers
	installmentCount := 0
	for _, flat := range allFlats {
		if flat.SaleDetail != nil {
			installmentCount = max(installmentCount, flat.SaleDetail.GetValidReceiptsCount())
		}
	}

	if installmentCount > 0 {
		installmentItems := make([]models.Header, 0, installmentCount)

		for i := 1; i <= installmentCount; i++ {
			installmentItems = append(installmentItems, models.Header{
				Heading: strconv.Itoa(i),
				Items: []models.Header{
					{Heading: "Number"},
					{Heading: "Date"},
					{Heading: "Amount", IsMonetary: true},
					{Heading: "Type"},
					{Heading: "CGST", IsMonetary: true},
					{Heading: "SGST", IsMonetary: true},
					{Heading: "Service Tax", IsMonetary: true},
					{Heading: "Swathch Bharat Cess", IsMonetary: true},
					{Heading: "Krishi Kalyan Cess", IsMonetary: true},
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

	// Create styles
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}

	numberStyle, err := createNumberStyle(file)
	if err != nil {
		return err
	}

	numberStyleYellow, err := createNumberStyleWithColor(file, "#FFFF00")
	if err != nil {
		return err
	}

	maxDepth := getMaxDepth(baseHeaders, 1)
	colIndex := 1

	headers := []models.Header{
		{Heading: "Member ID"},
	}
	headers = append(headers, baseHeaders...)

	_, err = renderHeaders(file, sheet, headers, 1, &colIndex, maxDepth, headerStyle)
	if err != nil {
		return err
	}

	// Set column widths
	for i := 1; i < colIndex; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		maxWidth := getMaxColumnWidth(file, sheet, colName, maxDepth)
		file.SetColWidth(sheet, colName, colName, maxWidth)
	}

	// Get monetary column indices
	monetaryColumns := getMonetaryColumnIndices(headers)
	yellowMonetaryColumns := getYellowMonetaryColumnIndices(headers)

	startRow := maxDepth + 1
	currentRow := startRow

	// Combine all active payment plan ratio items from all towers
	allActivePaymentPlanRatioItems := make([]models.TowerPaymentStatus, 0)
	for _, tower := range towers {
		allActivePaymentPlanRatioItems = append(allActivePaymentPlanRatioItems, tower.ActivePaymentPlanRatioItems...)
	}

	// Loop through ALL towers and their flats
	for _, tower := range towers {
		for i, flat := range tower.Flats {
			rowNum := currentRow
			values := flat.GetRowData(baseHeaders, tower.Name, models.SafePrint{
				ShouldPrint: i == 0 && tower.Name == "A",
			}, allActivePaymentPlanRatioItems)

			for colIdx, val := range values {
				colNum := colIdx + 1
				colName, _ := excelize.ColumnNumberToName(colNum)
				cell := fmt.Sprintf("%s%d", colName, rowNum)

				if monetaryColumns[colNum] {
					if numVal, ok := parseToFloat(val); ok {
						file.SetCellFloat(sheet, cell, numVal, -1, 64)

						if yellowMonetaryColumns[colNum] {
							file.SetCellStyle(sheet, cell, cell, numberStyleYellow)
						} else {
							file.SetCellStyle(sheet, cell, cell, numberStyle)
						}
					} else {
						file.SetCellValue(sheet, cell, val)
					}
				} else {
					file.SetCellValue(sheet, cell, val)
				}
			}
			currentRow++
		}
	}

	return nil
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

	// Create single sheet with all towers
	sheetErr := newMasterReportSheetAllTowers(reportFile, towerData)
	if sheetErr != nil {
		return nil, sheetErr
	}

	if err := reportFile.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := reportFile.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (s *reportService) generateMasterReport(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%s_master_report_%d.xlsx", fileNameBase, time.Now().Unix()),
	)
	w.Header().Set("Content-Length", fmt.Sprint(report.Len()))

	if _, err := w.Write(report.Bytes()); err != nil {
		payload.HandleError(w, err)
		return
	}
}