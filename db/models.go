package db

import "time"

// Disk represents a single disk attached to a VM.
type Disk struct {
	// Name is the name of the disk attached to a VM.
	Name string
	// Repo is the mount point of the repo.
	Repo string
}

// Snapshot holds information about a snapshot.
type Snapshot struct {
	ID        string    `storm:"id,unique,index"`
	VMID      string    `storm:"index"`
	CreatedAt time.Time `storm:"created_at"`
	Disks     []Disk
}
