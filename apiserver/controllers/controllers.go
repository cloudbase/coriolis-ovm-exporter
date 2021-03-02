package controllers

import (
	"coriolis-ovm-exporter/config"
	"coriolis-ovm-exporter/db"
	"net/http"

	"github.com/pkg/errors"
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

// APIController implements all API handlers.
type APIController struct {
	cfg *config.Config
	db  *db.Database
}

// LoginHandler attempts to authenticate against the OVM endpoint with the supplied credentials,
// and returns a JWT token.
func (a *APIController) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
