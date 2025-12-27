package model

import (
	"database/sql"
	"time"
)

// TicketType represents the type of ticket
type TicketType string

const (
	TicketTypeAppeal TicketType = "appeal" // 申诉
	TicketTypeReport TicketType = "report" // 举报
	TicketTypeOther  TicketType = "other"  // 其他
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"       // 待处理
	TicketStatusProcessing TicketStatus = "processing" // 处理中
	TicketStatusResolved   TicketStatus = "resolved"   // 已解决
	TicketStatusClosed     TicketStatus = "closed"     // 已关闭
)

// Ticket represents a ticket in the system
type Ticket struct {
	ID        int64         `db:"id" json:"id"`
	UserID    int64         `db:"user_id" json:"user_id"`
	PhotoID   sql.NullInt64 `db:"photo_id" json:"-"`
	Type      TicketType    `db:"type" json:"type"`
	Title     string        `db:"title" json:"title"`
	Content   string        `db:"content" json:"content"`
	Status    TicketStatus  `db:"status" json:"status"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt time.Time     `db:"updated_at" json:"updated_at"`
}

// TicketReply represents a reply to a ticket
type TicketReply struct {
	ID        int64     `db:"id" json:"id"`
	TicketID  int64     `db:"ticket_id" json:"ticket_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// TicketListItem represents a ticket in list view
type TicketListItem struct {
	ID        int64        `json:"id"`
	Type      TicketType   `json:"type"`
	Title     string       `json:"title"`
	Status    TicketStatus `json:"status"`
	PhotoID   *int64       `json:"photo_id,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

// TicketDetail represents detailed ticket information
type TicketDetail struct {
	ID        int64               `json:"id"`
	Type      TicketType          `json:"type"`
	Title     string              `json:"title"`
	Content   string              `json:"content"`
	Status    TicketStatus        `json:"status"`
	Photo     *TicketPhotoBrief   `json:"photo,omitempty"`
	Replies   []*TicketReplyItem  `json:"replies"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
}

// TicketPhotoBrief represents brief photo info for ticket
type TicketPhotoBrief struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// TicketReplyItem represents a reply in ticket detail
type TicketReplyItem struct {
	ID        int64            `json:"id"`
	Content   string           `json:"content"`
	User      *TicketUserBrief `json:"user"`
	CreatedAt string           `json:"created_at"`
}

// TicketUserBrief represents brief user info for ticket
type TicketUserBrief struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
}

// ToListItem converts Ticket to TicketListItem
func (t *Ticket) ToListItem() *TicketListItem {
	item := &TicketListItem{
		ID:        t.ID,
		Type:      t.Type,
		Title:     t.Title,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}

	if t.PhotoID.Valid {
		item.PhotoID = &t.PhotoID.Int64
	}

	return item
}

// ToDetail converts Ticket to TicketDetail
func (t *Ticket) ToDetail(photo *TicketPhotoBrief, replies []*TicketReplyItem) *TicketDetail {
	return &TicketDetail{
		ID:        t.ID,
		Type:      t.Type,
		Title:     t.Title,
		Content:   t.Content,
		Status:    t.Status,
		Photo:     photo,
		Replies:   replies,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

// ToReplyItem converts TicketReply to TicketReplyItem
func (r *TicketReply) ToReplyItem(user *TicketUserBrief) *TicketReplyItem {
	return &TicketReplyItem{
		ID:        r.ID,
		Content:   r.Content,
		User:      user,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}
}
