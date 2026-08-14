package tire

// Enumerations ported verbatim from Django's api/models/tire_model.py. The
// schema stores them as free varchar columns with no CHECK constraint, so
// these constants are the only place the vocabulary is written down.
const (
	StatusInStock  = "IN_STOCK"
	StatusMounted  = "MOUNTED"
	StatusRepair   = "REPAIR"
	StatusDisposed = "DISPOSED"
)

// Mount-log events. Removal is DISMOUNT, not REMOVE.
const (
	EventInstall  = "INSTALL"
	EventDismount = "DISMOUNT"
	EventRotation = "ROTATION"
)

// Assignment-request lifecycle.
const (
	RequestPending  = "PENDING"
	RequestApproved = "APPROVED"
	RequestRejected = "REJECTED"
)
