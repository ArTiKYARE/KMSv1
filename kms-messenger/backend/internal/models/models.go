package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	IsOnline  bool           `gorm:"default:false" json:"is_online"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Chat struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name,omitempty"` // For group chats
	Type      string         `gorm:"not null" json:"type"` // "direct" or "group"
	CreatedBy uint           `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Members   []ChatMember   `json:"members,omitempty"`
	Messages  []Message      `json:"messages,omitempty"`
}

type ChatMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ChatID    uint           `gorm:"not null;index" json:"chat_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Role      string         `gorm:"default:member" json:"role"` // "admin" or "member"
	JoinedAt  time.Time      `json:"joined_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	User      User           `json:"user"`
	Chat      Chat           `json:"chat"`
}

type Message struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ChatID    uint           `gorm:"not null;index" json:"chat_id"`
	SenderID  uint           `gorm:"not null;index" json:"sender_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	MessageType string       `gorm:"default:text" json:"message_type"` // "text", "file", "system"
	FileURL   string         `json:"file_url,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Sender    User           `json:"sender"`
	Chat      Chat           `json:"chat"`
}

type RefreshToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	User      User           `json:"user"`
}
