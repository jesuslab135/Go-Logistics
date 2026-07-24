package comment

import "time"

// Comment — port of api/models/comment_model.py:6
// Polymorphic via Django's GenericForeignKey, so ContentTypeID references the
// django_content_type table whose IDs Django assigns at migrate time. Any
// eventual Django decommission needs its own discriminator here.

type Comment struct {
	ID            int64
	CompanyID     int64
	ContentTypeID int64
	ObjectID      int32
	Body          string
	AuthorID      *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
