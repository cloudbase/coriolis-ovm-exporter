// Coriolis OVM exporter
// Copyright (C) 2021 Cloudbase Solutions SRL
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"coriolis-ovm-exporter/apiserver/params"
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

// DiskSnapshot represents a snapshot of a VM disk.
type DiskSnapshot struct {
	Name       string
	Repo       string
	SnapshotID string
	Path       string
	ParentPath string
	Chunks     []params.Chunk
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
