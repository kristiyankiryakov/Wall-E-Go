package handlers

import (
	"broker/internal/clients"
	"broker/internal/utils"
	"encoding/json"
	"net/http"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Authenticate(w http.ResponseWriter, r *http.Request)
}

type AuthHandlerImpl struct {
	authClient *clients.AuthClient
}

func NewAuthHandler(authClient *clients.AuthClient) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		authClient: authClient,
	}
}

func (h *AuthHandlerImpl) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Respond(w, http.StatusBadRequest, "invalid request", nil, err)
		return
	}

	token, err := h.authClient.RegisterUser(req.Username, req.Password)
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.Respond(
		w,
		http.StatusOK,
		"user registered successfully",
		map[string]string{
			"token": token,
		},
		nil,
	)
	return
}

func (h *AuthHandlerImpl) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Respond(w, http.StatusBadRequest, "invalid request", nil, err)
		return
	}

	token, err := h.authClient.Authenticate(req.Username, req.Password)
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.Respond(
		w,
		http.StatusOK,
		"user authenticated successfully",
		map[string]string{
			"token": token,
		},
		nil,
	)
	return
}
