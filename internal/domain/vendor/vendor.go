package vendor

import (
	"time"

	"github.com/shopspring/decimal"
)

// Vendor — port of api/models/vendor_model.py:4

type Vendor struct {
	ID                 int64
	CompanyID          int64
	Name               string
	IsMobileService    bool
	StreetAddress      string
	StreetAddressLine2 string
	City               string
	Region             string
	PostalCode         string
	Country            string
	Phone              string
	Website            string
	ContactName        string
	ContactPhone       string
	ContactEmail       string
	ExternalID         string
	Latitude           *decimal.Decimal
	Longitude          *decimal.Decimal
	IsFuelVendor       bool
	IsServiceVendor    bool
	IsPartsVendor      bool
	Labels             []string
	ArchivedAt         *time.Time
	CustomFields       map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
