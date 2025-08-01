package custom

//type DBType interface {
//	IsValid() bool
//}

// Custom types used in database models

type OrganizationStatus string
type UserRole string
type Facing string
type Salutation string
type Gender string
type MaritalStatus string
type Nationality string
type PreferenceLocationChargesType string
type PriceChargeType string
type PaymentPlanScope string
type PaymentPlanItemScope string
type PaymentPlanCondition string
type ReceiptMode string

const (
	ONLINE     ReceiptMode = "online"
	CASH       ReceiptMode = "cash"
	CHEQUE     ReceiptMode = "cheque"
	DD         ReceiptMode = "demand-draft"
	ADJUSTMENT ReceiptMode = "adjustment"
)

func (s ReceiptMode) IsValid() bool {
	switch s {
	case ONLINE, CASH, CHEQUE, DD, ADJUSTMENT:
		return true
	default:
		return false
	}
}

func (s ReceiptMode) RequireBankDetails() bool {
	switch s {
	case ONLINE, CHEQUE, DD:
		return true
	default:
		return false
	}
}

const (
	ACTIVE   OrganizationStatus = "active"
	INACTIVE OrganizationStatus = "inactive"
	ARCHIVE  OrganizationStatus = "archive"
)

func (s OrganizationStatus) IsValid() bool {
	switch s {
	case ACTIVE, INACTIVE, ARCHIVE:
		return true
	default:
		return false
	}
}

const (
	CIRCLEADMIN UserRole = "circle-admin"
	ORGADMIN    UserRole = "org-admin"
	ORGUSER     UserRole = "org-user"
	ORGVIEWER   UserRole = "org-viewer"
)

func (r UserRole) IsValid() bool {
	switch r {
	case CIRCLEADMIN, ORGADMIN, ORGUSER, ORGVIEWER:
		return true
	default:
		return false
	}
}

const (
	SPECIAL Facing = "Park/Road"
	DEFAULT Facing = "Default"
)

func (s Facing) IsValid() bool {
	switch s {
	case SPECIAL, DEFAULT:
		return true
	default:
		return false
	}
}

const (
	MR   Salutation = "Mr."
	MRS  Salutation = "Mrs."
	MISS Salutation = "Ms."
	DR   Salutation = "Dr."
	PROF Salutation = "Prof."
)

func (s Salutation) IsValid() bool {
	switch s {
	case MR, MRS, MISS, DR, PROF:
		return true
	default:
		return false
	}
}

const (
	MALE        Gender = "Male"
	FEMALE      Gender = "Female"
	TRANSGENDER Gender = "Transgender"
)

func (g Gender) IsValid() bool {
	switch g {
	case MALE, FEMALE, TRANSGENDER:
		return true
	default:
		return false
	}
}

const (
	MARRIED MaritalStatus = "Married"
	SINGLE  MaritalStatus = "Single"
)

func (m MaritalStatus) IsValid() bool {
	switch m {
	case MARRIED, SINGLE:
		return true
	default:
		return false
	}
}

const (
	RESIDENT Nationality = "Resident"
	PIO      Nationality = "PIO"
	NRI      Nationality = "NRI"
	OCI      Nationality = "OCI"
)

func (n Nationality) IsValid() bool {
	switch n {
	case RESIDENT, PIO, NRI, OCI:
		return true
	default:
		return false
	}
}

const (
	FLOOR  PreferenceLocationChargesType = "Floor"
	FACING PreferenceLocationChargesType = "Facing"
)

func (plc PreferenceLocationChargesType) IsValid() bool {
	switch plc {
	case FLOOR, FACING:
		return true
	default:
		return false
	}
}

const (
	PREFERENCELOCATIONCHARGE PriceChargeType = "location"
	OTHERCHARGE              PriceChargeType = "other"
)

func (plc PriceChargeType) IsValid() bool {
	switch plc {
	case PREFERENCELOCATIONCHARGE, OTHERCHARGE:
		return true
	default:
		return false
	}
}

const (
	DIRECT PaymentPlanScope = "Direct"
	TOWER  PaymentPlanScope = "Tower"
)

func (r PaymentPlanScope) IsValid() bool {
	switch r {
	case DIRECT, TOWER:
		return true
	default:
		return false
	}
}

const (
	ONBOOKING    PaymentPlanCondition = "on-booking"
	WITHINDAYS   PaymentPlanCondition = "within-days"
	ONALLOTMENT  PaymentPlanCondition = "on-allotment"
	ONTOWERSTAGE PaymentPlanCondition = "on-tower-stage"
	ONFlatSTAGE  PaymentPlanCondition = "on-flat-stage"
)

func (r PaymentPlanCondition) IsValid() bool {
	switch r {
	case ONBOOKING, WITHINDAYS, ONTOWERSTAGE, ONFlatSTAGE, ONALLOTMENT:
		return true
	default:
		return false
	}
}

const (
	SCOPE_SALE  PaymentPlanItemScope = "sale"
	SCOPE_TOWER PaymentPlanItemScope = "tower"
	SCOPE_FLAT  PaymentPlanItemScope = "flat"
)

func (r PaymentPlanItemScope) IsValid() bool {
	switch r {
	case SCOPE_FLAT, SCOPE_TOWER, SCOPE_SALE:
		return true
	default:
		return false
	}
}

var ValidPaymentPlanScopeCondtion = map[PaymentPlanItemScope][]PaymentPlanCondition{
	SCOPE_SALE: {
		ONBOOKING, WITHINDAYS, ONALLOTMENT,
	},
	SCOPE_TOWER: {
		ONTOWERSTAGE,
	},
	SCOPE_FLAT: {
		ONFlatSTAGE,
	},
}
