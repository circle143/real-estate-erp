package custom

// custom types used in database models

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

const (
	CIRCLEADMIN UserRole = "circle-admin"
	ORGADMIN    UserRole = "org-admin"
	ORGUSER     UserRole = "org-user"
)

const (
	DIRECT Seller = "direct"
	BROKER Seller = "broker"
	UNSOLD Seller = "unsold"
)

const (
	MR   Salutation = "mr"
	MRS  Salutation = "mrs"
	MISS Salutation = "miss"
)

const (
	MALE        Gender = "male"
	FEMALE      Gender = "female"
	TRANSGENDER Gender = "transgender"
)

const (
	MARRIED MaritalStatus = "married"
	SINGLE  MaritalStatus = "single"
)

const (
	RESIDENT Nationality = "resident"
	PIO      Nationality = "pio"
	NRI      Nationality = "nri"
	OCI      Nationality = "oci"
)