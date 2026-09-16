package handler

import (
	"encoding/json"
	"messanger/internal/middleware"
	"messanger/internal/service"
	"net/http"
)

type ChatHandler struct {
	service service.ChatService
}

func NewChatHandler(service service.ChatService) *ChatHandler {
	return &ChatHandler{
		service: service,
	}
}

type createChatResponse struct {
	ID string `json:"id"`
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
}

func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	chatID, err := h.service.CreateChat(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(createChatResponse{
		ID: chatID,
	})
}

func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chatID")

	if chatID == "" {
		writeError(w, http.StatusBadRequest, "chat id is required")
		return
	}

	var req addMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	if _, ok := middleware.UserID(r.Context()); !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	if err := h.service.AddMember(
		r.Context(),
		chatID,
		req.UserID,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
