package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Password  string         `gorm:"not null"`
	FirstName string         `gorm:"not null"`
	LastName  string         `gorm:"not null"`
	AvatarURL string
	IsActive  bool           `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type RefreshToken struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string         `gorm:"type:uuid;not null;index"`
	Token     string         `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time
	CreatedAt time.Time
	User      User           `gorm:"foreignKey:UserID"`
}

type Message struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChatID    string         `gorm:"type:uuid;not null;index"`
	SenderID  string         `gorm:"type:uuid;not null;index"`
	Content   string         `gorm:"type:text;not null"`
	MessageType string       `gorm:"default:'text'"` // text, file, system
	FileURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	Sender    User           `gorm:"foreignKey:SenderID"`
	Chat      Chat           `gorm:"foreignKey:ChatID"`
}

type Chat struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string
	Description string
	ChatType    string         `gorm:"default:'direct'"` // direct, group, channel
	CreatedBy   string         `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	
	Members     []ChatMember   `gorm:"foreignKey:ChatID"`
	Messages    []Message      `gorm:"foreignKey:ChatID"`
	Creator     User           `gorm:"foreignKey:CreatedBy"`
}

type ChatMember struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChatID    string         `gorm:"type:uuid;not null;index"`
	UserID    string         `gorm:"type:uuid;not null;index"`
	Role      string         `gorm:"default:'member'"` // admin, member
	JoinedAt  time.Time
	
	Chat      Chat           `gorm:"foreignKey:ChatID"`
	User      User           `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (Message) TableName() string {
	return "messages"
}

func (Chat) TableName() string {
	return "chats"
}

func (ChatMember) TableName() string {
	return "chat_members"
}
