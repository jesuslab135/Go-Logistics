package dto

import "time"

// NotificationResponse is the shape the bell icon will render once
// notifications are actually produced. The endpoint is a stub for now, so this
// type exists to pin the contract rather than to describe stored rows.
type NotificationResponse struct {
	ID        int64      `json:"id"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	URL       string     `json:"url"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type NotificationListResponse struct {
	Data        []NotificationResponse `json:"data"`
	UnreadCount int64                  `json:"unread_count"`
}
