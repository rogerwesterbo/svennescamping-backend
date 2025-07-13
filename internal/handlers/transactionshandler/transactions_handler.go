package transactionshandler

import (
	"net/http"
	"strconv"

	"github.com/rogerwesterbo/svennescamping-backend/internal/services"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/httphelpers"
)

func TransactionsHandler(transactionService *services.TransactionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get limit from query parameter, default to 25
		limitStr := r.URL.Query().Get("limit")
		limit := consts.TRANSACTION_LIMIT_DEFAULT

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
				limit = parsedLimit
			}
		}

		transactions, err := transactionService.GetTransactions(ctx, limit)
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch transactions")
			return
		}

		err = httphelpers.RespondWithJSON(w, http.StatusOK, transactions)
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to respond with transactions")
			return
		}
	}
}

func TransactionByIDHandler(transactionService *services.TransactionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract ID from URL path (you might want to use a router like gorilla/mux for this)
		id := r.URL.Query().Get("id")
		if id == "" {
			httphelpers.RespondWithError(w, http.StatusBadRequest, "Transaction ID is required")
			return
		}

		transaction, err := transactionService.GetTransactionByID(ctx, id)
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch transaction")
			return
		}

		err = httphelpers.RespondWithJSON(w, http.StatusOK, transaction)
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to respond with transaction")
			return
		}
	}
}

func RefreshCacheHandler(transactionService *services.TransactionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := transactionService.RefreshCache(ctx)
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to refresh cache")
			return
		}

		err = httphelpers.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Cache refreshed successfully"})
		if err != nil {
			httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to respond")
			return
		}
	}
}
