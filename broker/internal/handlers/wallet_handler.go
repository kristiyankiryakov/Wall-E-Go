package handlers

import (
	"broker/internal/clients"
	"broker/internal/models"
	"broker/internal/utils"
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
)

type WalletHandler interface {
	CreateWallet(w http.ResponseWriter, r *http.Request)
	ViewBalance(w http.ResponseWriter, r *http.Request)
	HealthCheck(w http.ResponseWriter, r *http.Request)
}

type WalletHandlerImpl struct {
	walletClient *clients.WalletClient
}

func NewWalletHandler(walletClient *clients.WalletClient) *WalletHandlerImpl {
	return &WalletHandlerImpl{
		walletClient: walletClient,
	}
}

func (h *WalletHandlerImpl) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Respond(w, http.StatusBadRequest, "invalid request", nil, err)
		return
	}

	walletID, err := h.walletClient.CreateWallet(r.Context(), req.Name)
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.Respond(
		w,
		http.StatusOK,
		"wallet created successfully",
		map[string]string{
			"walletID": walletID,
		},
		nil,
	)
	return
}

func (h *WalletHandlerImpl) ViewBalance(w http.ResponseWriter, r *http.Request) {
	walletID := r.URL.Query().Get("walletID")
	if walletID == "" {
		utils.Respond(w, http.StatusBadRequest, "missing walletID", nil, errors.New("missing walletID"))
		return
	}

	response, err := h.walletClient.ViewBalance(r.Context(), walletID)
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.Respond(
		w,
		http.StatusOK,
		"wallet balance retrieved successfully",
		models.ViewBalanceResponse{
			Name:    response.Name,
			Balance: response.Balance,
		},
		nil,
	)
	return
}

func (h *WalletHandlerImpl) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.walletClient.HealthCheck(ctx, &emptypb.Empty{})
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.Respond(w, http.StatusOK, "Wallet service is healthy", nil, nil)
}
