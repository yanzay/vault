package http

import (
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/hashicorp/vault/vault"
)

func handleSysSeal(core *vault.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			respondError(w, http.StatusMethodNotAllowed, nil)
			return
		}

		if err := core.Seal(); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}

		respondOk(w, nil)
	})
}

func handleSysUnseal(core *vault.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			respondError(w, http.StatusMethodNotAllowed, nil)
			return
		}

		// Parse the request
		var req UnsealRequest
		if err := parseRequest(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}
		if req.Key == "" {
			respondError(
				w, http.StatusBadRequest,
				errors.New("'key' must specified in request body as JSON"))
			return
		}

		// Decode the key, which is hex encoded
		key, err := hex.DecodeString(req.Key)
		if err != nil {
			respondError(
				w, http.StatusBadRequest,
				errors.New("'key' must be a valid hex-string"))
			return
		}

		// Attempt the unseal
		if _, err := core.Unseal(key); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}

		// Return the seal status
		handleSysSealStatusRaw(core, w, r)
	})
}

func handleSysSealStatus(core *vault.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			respondError(w, http.StatusMethodNotAllowed, nil)
			return
		}

		handleSysSealStatusRaw(core, w, r)
	})
}

func handleSysSealStatusRaw(core *vault.Core, w http.ResponseWriter, r *http.Request) {
	sealed, err := core.Sealed()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	sealConfig, err := core.SealConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	respondOk(w, &SealStatusResponse{
		Sealed:   sealed,
		T:        sealConfig.SecretThreshold,
		N:        sealConfig.SecretShares,
		Progress: core.SecretProgress(),
	})
}

type SealStatusResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

type UnsealRequest struct {
	Key string
}