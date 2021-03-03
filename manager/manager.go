package manager

import (
	"coriolis-ovm-exporter/apiserver/params"
	"coriolis-ovm-exporter/config"
	"coriolis-ovm-exporter/db"
	gErrors "coriolis-ovm-exporter/errors"
	"coriolis-ovm-exporter/internal"
	"log"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

// NewManager returns a new instance of SnapshotManager
func NewManager(cfg *config.Config) (*SnapshotManager, error) {
	db, err := db.NewDatabase(cfg.DBFile)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}
	return &SnapshotManager{
		db: db,
	}, nil
}

// SnapshotManager manages all snapshotting operations.
type SnapshotManager struct {
	db *db.Database
}

func (s *SnapshotManager) fetchVMSnapshotIDs(vmid string) ([]string, error) {
	snapshots, err := s.db.ListSnapshots(vmid)
	if err != nil {
		return nil, errors.Wrap(err, "fetching snapshots")
	}

	ret := make([]string, len(snapshots))
	for idx, snapshot := range snapshots {
		ret[idx] = snapshot.ID
	}
	return ret, nil
}

func (s *SnapshotManager) vmToParamsVirtualMachine(vm internal.VMConfig) (params.VirtualMachine, error) {
	vmDisks, err := vm.Disks()
	if err != nil {
		return params.VirtualMachine{}, errors.Wrap(err, "fetching VM disks")
	}
	disks := make([]params.Disk, len(vmDisks))
	for dIdx, disk := range vmDisks {
		disks[dIdx] = params.Disk{
			Name:       disk.Name,
			Path:       disk.Path,
			DeviceName: disk.DeviceName,
			Mode:       disk.Mode,
		}
	}

	snapshots, err := s.fetchVMSnapshotIDs(vm.Name)
	if err != nil {
		return params.VirtualMachine{}, errors.Wrap(err, "fetching snapshots")
	}
	return params.VirtualMachine{
		Name:         vm.Name,
		FriendlyName: vm.OVMSimpleName,
		UUID:         vm.UUID,
		Disks:        disks,
		Snapshots:    snapshots,
	}, nil
}

// ListVirtualMachines lists all virtual machines on this host.
func (s *SnapshotManager) ListVirtualMachines() ([]params.VirtualMachine, error) {
	vms, err := internal.ListAllVMs()
	if err != nil {
		return nil, errors.Wrap(err, "listing vms")
	}

	ret := make([]params.VirtualMachine, len(vms))
	for idx, vm := range vms {
		vmParams, err := s.vmToParamsVirtualMachine(vm)
		if err != nil {
			return nil, errors.Wrap(err, "fetching VM info")
		}

		ret[idx] = vmParams
	}
	return ret, nil
}

// GetVirtualMachine fetches information about a single virtual machine.
func (s *SnapshotManager) GetVirtualMachine(vmid string) (params.VirtualMachine, error) {
	vm, err := internal.GetVM(vmid)
	if err != nil {
		return params.VirtualMachine{}, errors.Wrap(err, "fetching VM info")
	}

	vmParams, err := s.vmToParamsVirtualMachine(vm)
	if err != nil {
		return params.VirtualMachine{}, errors.Wrap(err, "fetching VM params")
	}
	return vmParams, nil
}

func (s *SnapshotManager) snapshotToParamsSnapshot(snapshot internal.Snapshot) params.VMSnapshot {
	disks := make([]params.DiskSnapshot, len(snapshot.Disks))

	for idx, disk := range snapshot.Disks {
		disks[idx] = params.DiskSnapshot{
			Path:       disk.Path,
			ParentPath: disk.ParentPath,
			SnapshotID: disk.SnapshotID,
			Chunks:     disk.Chunks,
			Name:       disk.Name,
			Repo:       disk.Repo,
		}
	}
	ret := params.VMSnapshot{
		ID:    snapshot.SnapshotID,
		VMID:  snapshot.VMID,
		Disks: disks,
	}

	return ret
}

// CreateSnapshot creates a new snapshot of all VM disks.
func (s *SnapshotManager) CreateSnapshot(vmid string) (snap params.VMSnapshot, err error) {
	vm, err := internal.GetVM(vmid)
	if err != nil {
		return params.VMSnapshot{}, errors.Wrap(err, "fetching VM info")
	}

	snapshot, err := vm.CreateSnapshot(false)
	if err != nil {
		return params.VMSnapshot{}, errors.Wrap(err, "creating VM snapshot")
	}

	defer func() {
		if err != nil {
			err2 := snapshot.Delete()
			if err2 != nil {
				log.Printf("failed to cleanup snapshot: %q", err2)
			}
		}
	}()

	snapshotParams := s.snapshotToParamsSnapshot(snapshot)
	_, err = s.db.CreateSnapshot(snapshot.SnapshotID, vmid, snapshotParams.Disks)
	if err != nil {
		log.Printf("failed to save to db: %q", err)
		return params.VMSnapshot{}, err
	}
	return snapshotParams, nil
}

func (s *SnapshotManager) dbSnapToParamsSnapshots(snap db.Snapshot, squashChunks bool) params.VMSnapshot {
	var disks []params.DiskSnapshot
	if squashChunks {
		disks = make([]params.DiskSnapshot, len(snap.Disks))
		for idx, disk := range snap.Disks {
			disks[idx] = params.DiskSnapshot{
				ParentPath: disk.ParentPath,
				Path:       disk.Path,
				SnapshotID: disk.SnapshotID,
				Chunks:     internal.SquashChunks(disk.Chunks),
				Name:       disk.Name,
				Repo:       disk.Repo,
			}
		}
	} else {
		disks = snap.Disks
	}
	return params.VMSnapshot{
		ID:   snap.ID,
		VMID: snap.VMID,

		Disks: disks,
	}
}

func (s *SnapshotManager) getSnapshot(vmID, snapID string) (db.Snapshot, error) {
	snap, err := s.db.GetSnapshot(snapID)
	if err != nil {
		return db.Snapshot{}, errors.Wrap(err, "fetching snapshot")
	}
	if snap.VMID != vmID {
		return db.Snapshot{}, gErrors.NewConflictError("VM id missmatch")
	}
	return snap, nil
}

// GetSnapshot fetches information about a snapshot.
func (s *SnapshotManager) GetSnapshot(vmID, snapID string, squashChunks bool) (params.VMSnapshot, error) {
	snap, err := s.getSnapshot(vmID, snapID)
	if err != nil {
		return params.VMSnapshot{}, errors.Wrap(err, "fetching snapshot")
	}
	return s.dbSnapToParamsSnapshots(snap, squashChunks), nil
}

func (s *SnapshotManager) dbSnapToInternalSnap(snap db.Snapshot) internal.Snapshot {
	disks := make([]internal.DiskSnapshot, len(snap.Disks))

	for idx, disk := range snap.Disks {
		disks[idx] = internal.DiskSnapshot{
			Name:       disk.Name,
			Path:       disk.Path,
			SnapshotID: disk.SnapshotID,
			ParentPath: disk.ParentPath,
			Repo:       disk.Repo,
			Chunks:     disk.Chunks,
		}
	}
	ret := internal.Snapshot{
		SnapshotID: snap.ID,
		VMID:       snap.VMID,
		Disks:      disks,
	}
	return ret
}

// DeleteSnapshot deletes a single snapshot.
func (s *SnapshotManager) DeleteSnapshot(vmID, snapID string) error {
	snap, err := s.getSnapshot(vmID, snapID)
	if err != nil {
		errCause := errors.Cause(err)
		if errCause != storm.ErrNotFound {
			return errors.Wrap(err, "fetching snapshot")
		}
		return nil
	}

	internalSnap := s.dbSnapToInternalSnap(snap)
	if err := internalSnap.Delete(); err != nil {
		return err
	}

	err = s.db.DeleteSnapshot(snap.ID)
	if err != nil {
		return err
	}

	return nil
}
