package custom

// Custom types used in database models

type OrganizationStatus string
type UserRole string
type Seller string
type Salutation string
type Gender string
type MaritalStatus string
type Nationality string

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
)

func (r UserRole) IsValid() bool {
	switch r {
	case CIRCLEADMIN, ORGADMIN, ORGUSER:
		return true
	default:
		return false
	}
}

const (
	DIRECT Seller = "direct"
	BROKER Seller = "broker"
	UNSOLD Seller = "unsold"
)

func (s Seller) IsValid() bool {
	switch s {
	case DIRECT, BROKER, UNSOLD:
		return true
	default:
		return false
	}
}

const (
	MR   Salutation = "mr"
	MRS  Salutation = "mrs"
	MISS Salutation = "miss"
)

func (s Salutation) IsValid() bool {
	switch s {
	case MR, MRS, MISS:
		return true
	default:
		return false
	}
}

const (
	MALE        Gender = "male"
	FEMALE      Gender = "female"
	TRANSGENDER Gender = "transgender"
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
	MARRIED MaritalStatus = "married"
	SINGLE  MaritalStatus = "single"
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
	RESIDENT Nationality = "resident"
	PIO      Nationality = "pio"
	NRI      Nationality = "nri"
	OCI      Nationality = "oci"
)

func (n Nationality) IsValid() bool {
	switch n {
	case RESIDENT, PIO, NRI, OCI:
		return true
	default:
		return false
	}
}
