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

package params

var (
	// NotFoundResponse is returned when a resource is not found
	NotFoundResponse = APIErrorResponse{
		Error:   "Not Found",
		Details: "The resource you are looking for was not found",
	}
	// UnauthorizedResponse is a canned response for unauthorized access
	UnauthorizedResponse = APIErrorResponse{
		Error:   "Not Authorized",
		Details: "You do not have the required permissions to access this resource",
	}
)

// LoginResponse is the response clients get on successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse holds any errors generated during
// a request
type ErrorResponse struct {
	Errors map[string]string
}

// APIErrorResponse holds information about an error, returned by the API
type APIErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

// Chunk holds information about an extent.
type Chunk struct {
	Start  uint64 `json:"start"`
	Length uint64 `json:"length"`
	// Physical is the physical location on disk where this chunk resides.
	// When creating a reflink copy (copy-on-write) of a file,
	// if bytes get written to an extent in the original file, the
	// filesystem will write those bytes to a different physical location
	// on disk, ensuring that each copy of the file has it's own private
	// copy of the extent. When comparing differences between two copies
	// we'll be looking at the physical locations of the extents.
	Physical uint64 `json:"physical_start"`
}

// DiskSnapshot is a point in time snapshot of a disk.
type DiskSnapshot struct {
	ParentPath string  `json:"parent_path"`
	Path       string  `json:"path"`
	SnapshotID string  `json:"snapshot_id"`
	Chunks     []Chunk `json:"chunks"`
	Name       string  `json:"name"`
	Repo       string  `json:"repo_mountpoint"`
}

// VMSnapshot holds information about a single snapshot.
type VMSnapshot struct {
	ID   string `json:"id"`
	VMID string `json:"vm_id"`

	Disks []DiskSnapshot `json:"disks"`
}

// Disk holds information of a single disk attached to a VM.
type Disk struct {
	Name               string `json:"name"`
	Path               string `json:"path"`
	DeviceName         string `json:"device_name"`
	SnapshotCompatible bool   `json:"snapshot_compatible"`
	Mode               string `json:"mode"`
}

// VirtualMachine holds information about a single VM.
type VirtualMachine struct {
	FriendlyName       string `json:"friendly_name"`
	Name               string `json:"name"`
	UUID               string `json:"uuid"`
	Disks              []Disk `json:"disks"`
	SnapshotCompatible bool   `json:"snapshot_compatible"`
	// Snapshots is a list of snapshot IDs as fetched
	// from the database
	Snapshots []string `json:"snapshots"`
}
