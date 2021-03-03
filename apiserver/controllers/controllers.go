package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"coriolis-ovm-exporter/apiserver/auth"
	"coriolis-ovm-exporter/apiserver/params"
	"coriolis-ovm-exporter/config"
	gErrors "coriolis-ovm-exporter/errors"
	"coriolis-ovm-exporter/manager"
)

// NewAPIController returns a new instance of APIController
func NewAPIController(cfg *config.Config) (*APIController, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating config")
	}

	mgr, err := manager.NewManager(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}
	return &APIController{
		cfg: cfg,
		mgr: mgr,
	}, nil
}

func handleError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")
	origErr := errors.Cause(err)
	apiErr := params.APIErrorResponse{
		Details: origErr.Error(),
	}

	switch origErr.(type) {
	case *gErrors.NotFoundError:
		w.WriteHeader(http.StatusNotFound)
		apiErr.Error = "Not Found"
	case *gErrors.UnauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)
		apiErr.Error = "Not Authorized"
	case *gErrors.BadRequestError:
		w.WriteHeader(http.StatusBadRequest)
		apiErr.Error = "Bad Request"
	case *gErrors.ConflictError:
		w.WriteHeader(http.StatusConflict)
		apiErr.Error = "Conflict"
	default:
		w.WriteHeader(http.StatusInternalServerError)
		apiErr.Error = "Server error"
	}

	json.NewEncoder(w).Encode(apiErr)
	return
}

// APIController implements all API handlers.
type APIController struct {
	cfg *config.Config
	mgr *manager.SnapshotManager
}

// LoginHandler attempts to authenticate against the OVM endpoint with the supplied credentials,
// and returns a JWT token.
func (a *APIController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginInfo params.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginInfo); err != nil {
		handleError(w, gErrors.ErrBadRequest)
		return
	}

	cli := auth.NewOVMClient(loginInfo.Username, loginInfo.Password, a.cfg.OVMEndpoint)

	if err := cli.AttemptRequest(); err != nil {
		handleError(w, err)
		return
	}

	expireToken := time.Now().Add(a.cfg.JWTAuth.TimeToLive.Duration).Unix()
	claims := auth.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "gopherbin",
		},
		User: loginInfo.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.cfg.JWTAuth.Secret))
	if err != nil {
		handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params.LoginResponse{Token: tokenString})
}

// ListVMsHandler lists all VMs from all repositories on the system.
func (a *APIController) ListVMsHandler(w http.ResponseWriter, r *http.Request) {
	vms, err := a.mgr.ListVirtualMachines()
	if err != nil {
		handleError(w, err)
		return
	}
	json.NewEncoder(w).Encode(vms)
}

// GetVMHandler gets information about a single VM.
func (a *APIController) GetVMHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	vmInfo, err := a.mgr.GetVirtualMachine(vmID)
	if err != nil {
		handleError(w, err)
		return
	}
	json.NewEncoder(w).Encode(vmInfo)
}

// ListSnapshotsHandler lists all snapshots for a VM.
func (a *APIController) ListSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	snaps, err := a.mgr.ListSnapshots(vmID)
	if err != nil {
		handleError(w, err)
		return
	}
	json.NewEncoder(w).Encode(snaps)
}

// GetSnapshotHandler gets information about a single snapshot for a VM. It takes an optional
// query arg diff, which allows comparison of current snapshot, with a previous snapshot.
// The snapshot we are comparing to must exist and must be older than the current one.
func (a *APIController) GetSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	snapID, ok := vars["snapshotID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	squashChunksParam := r.URL.Query().Get("squashChunks")
	var squashChunks bool
	if squashChunksParam == "" {
		// Default to true
		squashChunks = true
	} else {
		squashChunks, _ = strconv.ParseBool(squashChunksParam)
	}

	compareTo := r.URL.Query().Get("compareTo")
	snapshot, err := a.mgr.GetSnapshot(vmID, snapID, compareTo, squashChunks)
	if err != nil {
		handleError(w, err)
		return
	}
	json.NewEncoder(w).Encode(snapshot)
}

// DeleteSnapshotHandler removes one snapshot associated with a VM.
func (a *APIController) DeleteSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	snapID, ok := vars["snapshotID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := a.mgr.DeleteSnapshot(vmID, snapID)
	if err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PurgeSnapshotsHandler deletes all snapshots associated with a VM.
func (a *APIController) PurgeSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := a.mgr.PurgeSnapshots(vmID); err != nil {
		handleError(w, err)
	}
	w.WriteHeader(http.StatusOK)
}

// CreateSnapshotHandler creates a snapshots for a VM.
func (a *APIController) CreateSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	// CreateSnapshot
	vars := mux.Vars(r)
	vmID, ok := vars["vmID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	snapData, err := a.mgr.CreateSnapshot(vmID)
	if err != nil {
		handleError(w, err)
		return
	}
	json.NewEncoder(w).Encode(snapData)
}

// ConsumeSnapshotHandler allows the caller to download arbitrary ranges of disk data from a
// disk snapshot.
func (a *APIController) ConsumeSnapshotHandler(w http.ResponseWriter, r *http.Request) {
}
