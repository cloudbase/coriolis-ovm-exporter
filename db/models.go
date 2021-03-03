package db

import (
	"coriolis-ovm-exporter/apiserver/params"
	"time"
)

// Snapshot holds information about a snapshot.
type Snapshot struct {
	ID        string `storm:"id,unique,index"`
	VMID      string `storm:"index"`
	CreatedAt time.Time
	Disks     []params.DiskSnapshot
}
