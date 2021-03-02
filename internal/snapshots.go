package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	// SnapshotDir is the folder relative to the repository
	// mount point where coriolis saves reflinked snapshots.
	SnapshotDir = "CoriolisSnapshots"
)

// Chunk holds information about an extent.
type Chunk struct {
	Start  uint64
	Length uint64
	// Physical is the physical location on disk where this chunk resides.
	// When creating a reflink copy (copy-on-write) of a file,
	// if bytes get written to an extent in the original file, the
	// filesystem will write those bytes to a different physical location
	// on disk, ensuring that each copy of the file has it's own private
	// copy of the extent. When comparing differences between two copies
	// we'll be looking at the physical locations of the extents.
	Physical uint64
}

// DiskSnapshot represents a snapshot of a VM disk.
type DiskSnapshot struct {
	Name       string
	Repo       string
	SnapshotID string
	Chunks     []Chunk
}

// DeleteSnapshot deletes files associated with this disk snapshot.
func (d DiskSnapshot) DeleteSnapshot() error {
	snapshotDir := filepath.Join(d.Repo, SnapshotDir, d.SnapshotID)
	if _, err := os.Stat(snapshotDir); err != nil {
		if os.IsNotExist(err) == false {
			return errors.Wrap(err, "accessing snapshot dir")
		}
		// Snapshot dir is gone. Desired state equals actual state.
		return nil
	}

	diskSnap := filepath.Join(snapshotDir, d.Name)
	if err := os.Remove(diskSnap); err != nil {
		return errors.Wrap(err, "removing snapshot")
	}

	contents, err := ioutil.ReadDir(snapshotDir)
	if err != nil {
		return errors.Wrap(err, "accessing snapshot dir")
	}

	if len(contents) == 0 {
		// There are no more snapshots in this folder.
		// Cleanup empty snapshot dir.
		// TODO: This might cause a race condition. Investigate
		// if we need to add some locking for remove operations.
		os.RemoveAll(snapshotDir)
	}
	return nil
}

// Snapshot represents a snapshot in time of the disks of a VM.
type Snapshot struct {
	SnapshotID string
	VMID       string

	Disks []DiskSnapshot
}

// Delete removes all associated disk snapshots.
func (s Snapshot) Delete() error {
	for _, disk := range s.Disks {
		if err := disk.DeleteSnapshot(); err != nil {
			return errors.Wrap(err, "deleting disk snapshot")
		}
	}

	return nil
}

// GetSnapshot gets a new Snapshot instance from an ID.
func GetSnapshot(snapID string) (Snapshot, error) {
	return Snapshot{}, fmt.Errorf("not implemented")
}
