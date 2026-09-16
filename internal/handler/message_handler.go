package handler

import (
	"encoding/json"
	"errors"
	"messanger/internal/middleware"
	"messanger/internal/service"
	"net/http"
)

type MessageHandler struct {
	service service.MessageService
}

func NewMessageHandler(service service.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
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
