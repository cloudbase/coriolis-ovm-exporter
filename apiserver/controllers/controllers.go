package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"coriolis-ovm-exporter/apiserver/auth"
	"coriolis-ovm-exporter/apiserver/params"
	"coriolis-ovm-exporter/config"
	"coriolis-ovm-exporter/db"
	gErrors "coriolis-ovm-exporter/errors"
)

// NewAPIController returns a new instance of APIController
func NewAPIController(cfg *config.Config) (*APIController, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating config")
	}

	db, err := db.NewDatabase(cfg.DBFile)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}
	return &APIController{
		cfg: cfg,
		db:  db,
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
	db  *db.Database
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
}

// GetVMHandler gets information about a single VM.
func (a *APIController) GetVMHandler(w http.ResponseWriter, r *http.Request) {
}

// ListSnapshotsHandler lists all snapshots for a VM.
func (a *APIController) ListSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
}

// GetSnapshotHandler gets information about a single snapshot for a VM. It takes an optional
// query arg diff, which allows comparison of current snapshot, with a previous snapshot.
// The snapshot we are comparing to must exist and must be older than the current one.
func (a *APIController) GetSnapshotHandler(w http.ResponseWriter, r *http.Request) {
}

// DeleteSnapshotHandler removes one snapshot associated with a VM.
func (a *APIController) DeleteSnapshotHandler(w http.ResponseWriter, r *http.Request) {
}

// PurgeSnapshotsHandler deletes all snapshots associated with a VM.
func (a *APIController) PurgeSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
}

// CreateSnapshotHandler creates a snapshots for a VM.
func (a *APIController) CreateSnapshotHandler(w http.ResponseWriter, r *http.Request) {
}

// ConsumeSnapshotHandler allows the caller to download arbitrary ranges of disk data from a
// disk snapshot.
func (a *APIController) ConsumeSnapshotHandler(w http.ResponseWriter, r *http.Request) {
}
