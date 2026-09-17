package handler

import (
	"encoding/json"
	"errors"
	"messanger/internal/middleware"
	"messanger/internal/service"
	"net/http"
)

type ChatHandler struct {
	service     service.ChatService
	userService service.UserService
}

func NewChatHandler(service service.ChatService, userService service.UserService) *ChatHandler {
	return &ChatHandler{
		service:     service,
		userService: userService,
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

	callerID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
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

	err := h.service.AddMember(
		r.Context(),
		chatID,
		req.UserID,
		callerID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotChatMember):
			writeError(w, http.StatusForbidden, "you are not a member of this chat")
		default:
			writeError(w, http.StatusInternalServerError, "failed to add member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type chatListItem struct {
	ID           string `json:"id"`
	LastMessage  string `json:"last_message"`
	LastSenderID string `json:"last_sender_id"`
	LastMsgTime  string `json:"last_msg_time"`
	MemberCount  int    `json:"member_count"`
}

func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	chats, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	result := make([]chatListItem, 0, len(chats))
	for _, c := range chats {
		result = append(result, chatListItem{
			ID:           c.ID,
			LastMessage:  c.LastMessage,
			LastSenderID: c.LastSenderID,
			LastMsgTime:  c.LastMsgTime,
			MemberCount:  c.MemberCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

type memberItem struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (h *ChatHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chatID")
	if chatID == "" {
		writeError(w, http.StatusBadRequest, "chat id is required")
		return
	}

	callerID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	isMember, err := h.service.IsMember(r.Context(), chatID, callerID)
	if err != nil || !isMember {
		writeError(w, http.StatusForbidden, "you are not a member of this chat")
		return
	}

	memberIDs, err := h.service.GetMemberIDs(r.Context(), chatID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	result := make([]memberItem, 0, len(memberIDs))
	for _, uid := range memberIDs {
		username, _ := h.userService.GetByID(r.Context(), uid)
		if username == "" {
			username = uid[:8]
		}
		result = append(result, memberItem{UserID: uid, Username: username})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
