package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalcGST(t *testing.T) {
	value := decimal.NewFromInt(2000)

	rateFive := 5
	rateFiveAmountWant := decimal.NewFromFloat(1904.76)
	rateFiveGSTAmount := decimal.NewFromFloat(47.62)

	gstCalcFive := calcGST(value, rateFive)

	if !gstCalcFive.Amount.Equal(rateFiveAmountWant) ||
		!gstCalcFive.CGST.Equal(rateFiveGSTAmount) ||
		!gstCalcFive.SGST.Equal(rateFiveGSTAmount) {
		t.Errorf("want:\nAmount: %s, CGST: %s, SGST: %s\nGot: Amount: %s, CGST: %s, SGST: %s",
			rateFiveAmountWant.String(),
			rateFiveGSTAmount.String(),
			rateFiveGSTAmount.String(),
			gstCalcFive.Amount.String(),
			gstCalcFive.CGST.String(),
			gstCalcFive.SGST.String(),
		)
	}

	rateOne := 1
	rateOneAmountWant := decimal.NewFromFloat(1980.20)
	rateOneGSTAmount := decimal.NewFromFloat(9.9)

	gstCalcOne := calcGST(value, rateOne)

	if !gstCalcOne.Amount.Equal(rateOneAmountWant) ||
		!gstCalcOne.CGST.Equal(rateOneGSTAmount) ||
		!gstCalcOne.SGST.Equal(rateOneGSTAmount) {
		t.Errorf("want:\nAmount: %s, CGST: %s, SGST: %s\nGot: Amount: %s, CGST: %s, SGST: %s",
			rateOneAmountWant.String(),
			rateOneGSTAmount.String(),
			rateOneGSTAmount.String(),
			gstCalcOne.Amount.String(),
			gstCalcOne.CGST.String(),
			gstCalcOne.SGST.String(),
		)
	}

}
