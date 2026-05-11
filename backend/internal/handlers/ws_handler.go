package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kms-messenger/backend/internal/websocket"
)

type WSHandler struct {
	hub *websocket.Hub
}

func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{
		hub: hub,
	}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h.hub.HandleWebSocket(c.Writer, c.Request, userID.(string))
}
