package manager

import (
	"coriolis-ovm-exporter/apiserver/params"
	"coriolis-ovm-exporter/config"
	"coriolis-ovm-exporter/db"
	gErrors "coriolis-ovm-exporter/errors"
	"coriolis-ovm-exporter/internal"
	"fmt"
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

func (s *SnapshotManager) squashChunks(disks []params.DiskSnapshot) []params.DiskSnapshot {
	ret := make([]params.DiskSnapshot, len(disks))
	for idx, disk := range disks {
		ret[idx] = params.DiskSnapshot{
			ParentPath: disk.ParentPath,
			Path:       disk.Path,
			SnapshotID: disk.SnapshotID,
			Chunks:     internal.SquashChunks(disk.Chunks),
			Name:       disk.Name,
			Repo:       disk.Repo,
		}
	}
	return ret
}

func (s *SnapshotManager) dbSnapToParamsSnapshots(snap db.Snapshot, squashChunks bool) params.VMSnapshot {
	var disks []params.DiskSnapshot
	fmt.Println(squashChunks)
	if squashChunks == true {
		disks = s.squashChunks(snap.Disks)
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

func (s *SnapshotManager) compareChunks(first, second []params.Chunk) []params.Chunk {
	var ret []params.Chunk
	for _, val := range first {
		var found bool = false
		for _, prevVal := range second {
			if val.Physical == prevVal.Physical {
				if val.Start == prevVal.Start && val.Length == prevVal.Length {
					found = true
					break
				}
			}
		}
		// This is a new extent. Add it to the list.
		if !found {
			ret = append(ret, val)
		}
	}
	return ret
}

func (s *SnapshotManager) getDiffSnapshot(snap, compareTo db.Snapshot) (db.Snapshot, error) {
	if !compareTo.CreatedAt.Before(snap.CreatedAt) {
		return db.Snapshot{}, gErrors.NewBadRequestError(
			"compareTo snapshot must be older than this snapshot")
	}

	if snap.VMID != compareTo.VMID {
		return db.Snapshot{}, gErrors.NewBadRequestError(
			"compareTo snapshot does not belong to this VM")
	}

	newDisks := make([]params.DiskSnapshot, len(snap.Disks))

	for idx, disk := range snap.Disks {
		var chunks []params.Chunk = disk.Chunks
		for _, compareDisk := range compareTo.Disks {
			if compareDisk.Name == disk.Name {
				chunks = s.compareChunks(disk.Chunks, compareDisk.Chunks)
				break
			}
		}
		newDisks[idx] = params.DiskSnapshot{
			ParentPath: disk.ParentPath,
			Path:       disk.Path,
			SnapshotID: disk.SnapshotID,
			Chunks:     chunks,
			Name:       disk.Name,
			Repo:       disk.Repo,
		}
	}
	// TODO: should we copy the values?
	snap.Disks = newDisks
	return snap, nil
}

// GetSnapshot fetches information about a snapshot.
func (s *SnapshotManager) GetSnapshot(vmID, snapID, compareTo string, squashChunks bool) (params.VMSnapshot, error) {
	var snap db.Snapshot
	var err error

	requestedSnap, err := s.getSnapshot(vmID, snapID)
	if err != nil {
		return params.VMSnapshot{}, errors.Wrap(err, "fetching snapshot")
	}

	var compareToSnap db.Snapshot
	if compareTo != "" {
		compareToSnap, err = s.getSnapshot(vmID, compareTo)
		if err != nil {
			return params.VMSnapshot{}, err
		}
		snap, err = s.getDiffSnapshot(requestedSnap, compareToSnap)
		if err != nil {
			return params.VMSnapshot{}, err
		}
	} else {
		snap = requestedSnap
	}
	return s.dbSnapToParamsSnapshots(snap, squashChunks), nil
}

// ListSnapshots lists all snapshots for a VM
func (s *SnapshotManager) ListSnapshots(vmID string) ([]params.VMSnapshot, error) {
	vm, err := s.GetVirtualMachine(vmID)
	if err != nil {
		return nil, err
	}

	ret := make([]params.VMSnapshot, len(vm.Snapshots))
	for idx, snap := range vm.Snapshots {
		retSnap, err := s.GetSnapshot(vmID, snap, "", true)
		if err != nil {
			return nil, err
		}
		ret[idx] = retSnap
	}
	return ret, nil
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

// PurgeSnapshots deletes all snapshots for a VM.
func (s *SnapshotManager) PurgeSnapshots(vmID string) error {
	vm, err := s.GetVirtualMachine(vmID)
	if err != nil {
		return err
	}
	for _, snap := range vm.Snapshots {
		if err := s.DeleteSnapshot(vmID, snap); err != nil {
			return err
		}
	}
	return nil
}
