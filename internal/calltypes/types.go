package calltypes

import "time"

type Session struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenant_id"`
	RoomName  string     `json:"room_name"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type Turn struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}