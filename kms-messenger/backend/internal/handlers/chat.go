package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kms-backend/internal/models"
)

type ChatHandler struct {
	DB *gorm.DB
}

func NewChatHandler(db *gorm.DB) *ChatHandler {
	return &ChatHandler{DB: db}
}

type CreateChatInput struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type" binding:"required,oneof=direct group"`
	UserIDs []uint `json:"user_ids,omitempty"`
}

type ChatResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name,omitempty"`
	Type      string         `json:"type"`
	CreatedBy uint           `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	Members   []MemberResponse `json:"members"`
}

type MemberResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

type MessageInput struct {
	Content string `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID          uint        `json:"id"`
	ChatID      uint        `json:"chat_id"`
	SenderID    uint        `json:"sender_id"`
	Content     string      `json:"content"`
	MessageType string      `json:"message_type"`
	FileURL     string      `json:"file_url,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	Sender      UserSummary `json:"sender"`
}

type UserSummary struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	userID, _ := c.Get("userID").(uint)

	var chats []models.Chat
	if err := h.DB.Preload("Members.User").Where("id IN (SELECT chat_id FROM chat_members WHERE user_id = ?)", userID).Find(&chats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chats"})
		return
	}

	response := make([]ChatResponse, len(chats))
	for i, chat := range chats {
		members := make([]MemberResponse, len(chat.Members))
		for j, member := range chat.Members {
			members[j] = MemberResponse{
				ID:        member.User.ID,
				Email:     member.User.Email,
				FirstName: member.User.FirstName,
				LastName:  member.User.LastName,
				AvatarURL: member.User.AvatarURL,
				Role:      member.Role,
			}
		}
		response[i] = ChatResponse{
			ID:        chat.ID,
			Name:      chat.Name,
			Type:      chat.Type,
			CreatedBy: chat.CreatedBy,
			CreatedAt: chat.CreatedAt,
			Members:   members,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID, _ := c.Get("userID").(uint)

	var input CreateChatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For direct chat, check if chat already exists between users
	if input.Type == "direct" && len(input.UserIDs) == 1 {
		var existingChat models.Chat
		if err := h.DB.Joins("JOIN chat_members cm1 ON cm1.chat_id = chats.id").
			Joins("JOIN chat_members cm2 ON cm2.chat_id = chats.id").
			Where("chats.type = ?", "direct").
			Where("cm1.user_id = ? AND cm2.user_id = ?", userID, input.UserIDs[0]).
			First(&existingChat).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Direct chat already exists", "chat_id": existingChat.ID})
			return
		}
	}

	chat := models.Chat{
		Name:      input.Name,
		Type:      input.Type,
		CreatedBy: userID,
	}

	tx := h.DB.Begin()
	if err := tx.Create(&chat).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	// Add creator as admin
	creatorMember := models.ChatMember{
		ChatID: chat.ID,
		UserID: userID,
		Role:   "admin",
	}
	if err := tx.Create(&creatorMember).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator to chat"})
		return
	}

	// Add other members
	for _, memberID := range input.UserIDs {
		if memberID == userID {
			continue
		}
		member := models.ChatMember{
			ChatID: chat.ID,
			UserID: memberID,
			Role:   "member",
		}
		if err := tx.Create(&member).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to chat"})
			return
		}
	}

	tx.Commit()

	// Fetch complete chat with members
	var completeChat models.Chat
	if err := h.DB.Preload("Members.User").First(&completeChat, chat.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created chat"})
		return
	}

	members := make([]MemberResponse, len(completeChat.Members))
	for i, member := range completeChat.Members {
		members[i] = MemberResponse{
			ID:        member.User.ID,
			Email:     member.User.Email,
			FirstName: member.User.FirstName,
			LastName:  member.User.LastName,
			AvatarURL: member.User.AvatarURL,
			Role:      member.Role,
		}
	}

	c.JSON(http.StatusCreated, ChatResponse{
		ID:        completeChat.ID,
		Name:      completeChat.Name,
		Type:      completeChat.Type,
		CreatedBy: completeChat.CreatedBy,
		CreatedAt: completeChat.CreatedAt,
		Members:   members,
	})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatID := c.Param("id")
	userID, _ := c.Get("userID").(uint)

	// Verify user is member of chat
	var member models.ChatMember
	if err := h.DB.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var messages []models.Message
	if err := h.DB.Preload("Sender").Where("chat_id = ?", chatID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	response := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = MessageResponse{
			ID:          msg.ID,
			ChatID:      msg.ChatID,
			SenderID:    msg.SenderID,
			Content:     msg.Content,
			MessageType: msg.MessageType,
			FileURL:     msg.FileURL,
			CreatedAt:   msg.CreatedAt,
			Sender: UserSummary{
				ID:        msg.Sender.ID,
				FirstName: msg.Sender.FirstName,
				LastName:  msg.Sender.LastName,
				AvatarURL: msg.Sender.AvatarURL,
			},
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	chatID := c.Param("id")
	userID, _ := c.Get("userID").(uint)

	// Verify user is member of chat
	var member models.ChatMember
	if err := h.DB.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var input MessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := models.Message{
		ChatID:      uint(member.ChatID),
		SenderID:    userID,
		Content:     input.Content,
		MessageType: "text",
	}

	if err := h.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Fetch complete message with sender
	var completeMessage models.Message
	if err := h.DB.Preload("Sender").First(&completeMessage, message.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sent message"})
		return
	}

	c.JSON(http.StatusCreated, MessageResponse{
		ID:          completeMessage.ID,
		ChatID:      completeMessage.ChatID,
		SenderID:    completeMessage.SenderID,
		Content:     completeMessage.Content,
		MessageType: completeMessage.MessageType,
		FileURL:     completeMessage.FileURL,
		CreatedAt:   completeMessage.CreatedAt,
		Sender: UserSummary{
			ID:        completeMessage.Sender.ID,
			FirstName: completeMessage.Sender.FirstName,
			LastName:  completeMessage.Sender.LastName,
			AvatarURL: completeMessage.Sender.AvatarURL,
		},
	})
}
