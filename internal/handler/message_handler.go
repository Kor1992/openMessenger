package handler

import (
	"encoding/json"
	"errors"
	"messanger/internal/middleware"
	"messanger/internal/service"
	"net/http"
	"strconv"
)

type MessageHandler struct {
	service     service.MessageService
	chatService service.ChatService
}

func NewMessageHandler(service service.MessageService, chatService service.ChatService) *MessageHandler {
	return &MessageHandler{
		service:     service,
		chatService: chatService,
	}
}

type sendMessageRequest struct {
	Text string `json:"text"`
}

type sendMessageResponse struct {
	ID string `json:"id"`
}

func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chatID")

	if chatID == "" {
		writeError(w, http.StatusBadRequest, "chat id is required")
		return
	}

	senderID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req sendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	messageID, err := h.service.SendMessage(
		r.Context(),
		chatID,
		senderID,
		req.Text,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMessage):
			writeError(w, http.StatusBadRequest, "message text is required")
		case errors.Is(err, service.ErrNotChatMember):
			writeError(w, http.StatusForbidden, "user is not a chat member")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(sendMessageResponse{
		ID: messageID,
	})
}

type messageItem struct {
	ID        string `json:"id"`
	SenderID  string `json:"sender_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chatID")
	if chatID == "" {
		writeError(w, http.StatusBadRequest, "chat id is required")
		return
	}

	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	isMember, err := h.chatService.IsMember(r.Context(), chatID, userID)
	if err != nil || !isMember {
		writeError(w, http.StatusForbidden, "you are not a member of this chat")
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	msgs, err := h.service.ListByChat(r.Context(), chatID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	result := make([]messageItem, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, messageItem{
			ID:        m.ID,
			SenderID:  m.SenderID,
			Text:      m.Text,
			CreatedAt: m.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
