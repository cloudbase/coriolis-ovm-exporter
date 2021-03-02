package internal

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	// VirtualMachinesDir is the directory where VM config files
	// are stored.
	VirtualMachinesDir = "VirtualMachines"
)

// Disk represents one VM disk
type Disk struct {
	Name       string
	Path       string
	DeviceName string
	Mode       string
	Repo       Repo
	ObjectType string
}

// CanClone returns a boolean value indicating whether or not
// this disk can be reflinked.
func (d Disk) CanClone() bool {
	if d.Repo.Filesystem != "ocfs2" {
		log.Printf("Filesystem is not ocfs2: %s", d.Repo.Filesystem)
		return false
	}

	if d.ObjectType != "" && d.ObjectType != "VIRTUAL_DISK" {
		return false
	}

	return true
}

// CreateSnapshot creates a reflink copy of a virtual disk and returns
// a DiskSnapshot object.
func (d Disk) CreateSnapshot(snapID string) (snap DiskSnapshot, err error) {
	if d.CanClone() == false {
		return DiskSnapshot{}, fmt.Errorf("repository of %s does not support reflink cloning", d.Name)
	}

	snapshotDir := filepath.Join(d.Repo.MountPoint, SnapshotDir, snapID)
	if _, err := os.Stat(snapshotDir); err != nil {
		if os.IsNotExist(err) == false {
			return DiskSnapshot{}, fmt.Errorf("failed to stat %s", snapshotDir)
		}
		log.Printf("Creating snapshot dir: %s", snapshotDir)
		if err := os.MkdirAll(snapshotDir, 00750); err != nil {
			return DiskSnapshot{}, fmt.Errorf("failed to create %s", snapshotDir)
		}
	}
	snapFile := filepath.Join(snapshotDir, d.Name)

	defer func() {
		if err != nil {
			os.Remove(snapFile)
			if _, err := os.Stat(snapshotDir); err == nil {
				contents, err2 := ioutil.ReadDir(snapshotDir)
				if err2 != nil {
					return
				}

				if len(contents) == 0 {
					os.RemoveAll(snapshotDir)
				}
			}
		}
	}()

	log.Printf("Creating snap of %s to %s", d.Path, snapFile)

	if err := IOctlOCFS2Reflink(d.Path, snapFile); err != nil {
		return DiskSnapshot{}, errors.Wrap(err, "creating reflink")
	}

	chunks, err := getFileExtents(snapFile)
	if err != nil {
		return DiskSnapshot{}, err
	}

	snap = DiskSnapshot{
		Name:       d.Name,
		Repo:       d.Repo.MountPoint,
		SnapshotID: snapID,
		Chunks:     chunks,
	}
	return snap, nil
}

// VMConfig is a stripped down VM config, containing only
// the fields we care about.
type VMConfig struct {
	// OVM_simple_name is the friendly name for a VM
	OVMSimpleName string `toml:"OVM_simple_name"`
	// Name is the internal name of the VM. This is usually
	// just the UUID with the hyphens removed.
	Name string
	// UUID is the UUID of the VM.
	UUID string
	// Disk is a list of paths to virtual machine disks.
	DiskArray []string `toml:"disk"`
}

// CreateSnapshot creates a new snapshot and returns the ID of the snapshot.
func (v VMConfig) CreateSnapshot(shutdownVM bool) (snapshot Snapshot, err error) {
	snapID := uuid.NewString()

	if v.CanClone() == false {
		err = fmt.Errorf("VM does not support reflink cloning")
		return
	}

	disks, err := v.Disks()
	if err != nil {
		return
	}

	var snapDisks []DiskSnapshot

	defer func() {
		if err != nil {
			for _, disk := range snapDisks {
				disk.DeleteSnapshot()
			}
		}
	}()

	for _, disk := range disks {
		var snap DiskSnapshot
		snap, err = disk.CreateSnapshot(snapID)
		if err != nil {
			err = errors.Wrap(err, "creating disk snapshot")
			return
		}
		snapDisks = append(snapDisks, snap)
	}

	snapshot.Disks = snapDisks
	snapshot.SnapshotID = snapID
	snapshot.VMID = v.Name

	return
}

// CanClone returns true if all disks attached to this instance are
// cloneable.
func (v VMConfig) CanClone() bool {
	disks, err := v.Disks()
	if err != nil {
		return false
	}

	for _, disk := range disks {
		if disk.CanClone() == false {
			return false
		}
	}

	return true
}

// Disks returns an array of Disk objects, representing the
// disks attached to a VM.
func (v VMConfig) Disks() ([]Disk, error) {
	var ret []Disk

	repos, err := ParseRepos()
	if err != nil {
		return ret, err
	}

	for _, val := range v.DiskArray {
		schemaSplit := strings.SplitN(val, ":", 2)
		if len(schemaSplit) != 2 || schemaSplit[0] != "file" {
			log.Printf("ignoring non file disk: %s", val)
			continue
		}

		details := strings.Split(schemaSplit[1], ",")
		if len(details) != 3 {
			// expecting path,device_name,mode
			log.Printf("unexpected number of values (%d) for %s (expected 3)", len(details), val)
			continue
		}

		baseName := filepath.Base(details[0])
		dsk := Disk{
			Path:       details[0],
			DeviceName: details[1],
			Mode:       details[2],
			Name:       baseName,
		}

		for _, repo := range repos {
			meta, err := repo.Meta()
			if err != nil {
				log.Printf("failed to find repo metadata: %q", err)
			} else {
				if diskMeta, ok := meta[dsk.Name]; ok {
					dsk.ObjectType = diskMeta.ObjectType
					dsk.Repo = repo
					// No need to attempt prefix match.
					break
				}
			}
			// attempt a prefix match. Not all disks are present in
			// repo metadata.
			if strings.HasPrefix(dsk.Path, repo.MountPoint) {
				dsk.Repo = repo
				break
			}
		}

		ret = append(ret, dsk)
	}
	return ret, nil
}

func pruneConfig(cfgFile string) (string, error) {
	lookFor := map[string]int{
		"OVM_simple_name": 1,
		"disk":            1,
		"uuid":            1,
		"name":            1,
	}

	file, err := os.Open(cfgFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var ret []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		elems := strings.Split(line, "=")
		if len(elems) == 2 {
			if _, ok := lookFor[strings.Trim(elems[0], " ")]; ok {
				ret = append(ret, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(ret, "\n"), nil
}

// ParseVMConfig returns a new Config
func ParseVMConfig(cfgFile string) (VMConfig, error) {
	cfg, err := pruneConfig(cfgFile)
	if err != nil {
		return VMConfig{}, err
	}

	var config VMConfig
	if _, err := toml.Decode(cfg, &config); err != nil {
		return config, err
	}
	return config, nil
}

// ListAllVMs returns a list of VMConfig from all currently known
// repositories.
func ListAllVMs() ([]VMConfig, error) {
	repos, err := ParseRepos()
	if err != nil {
		return nil, errors.Wrap(err, "fetching repos")
	}

	var ret []VMConfig

	for _, repo := range repos {
		vms, err := ListVMs(repo)
		if err != nil {
			return nil, errors.Wrap(err, "listing vms in repo")
		}
		ret = append(ret, vms...)
	}
	return ret, nil
}

// ListVMs returns a list of VMConfig from a repository.
func ListVMs(repo Repo) ([]VMConfig, error) {
	var ret []VMConfig

	vmDirPath := filepath.Join(repo.MountPoint, VirtualMachinesDir)

	if _, err := os.Stat(vmDirPath); err != nil {
		return ret, errors.Wrap(err, "accessing repo VM dir")
	}

	vmDirs, err := ioutil.ReadDir(vmDirPath)
	if err != nil {
		return ret, errors.Wrap(err, "listing VM dir")
	}

	for _, item := range vmDirs {
		if item.IsDir() == false {
			continue
		}

		vmCfgFile := filepath.Join(vmDirPath, item.Name(), "vm.cfg")
		if _, err := os.Stat(vmCfgFile); err != nil {
			continue
		}
		vmCfg, err := ParseVMConfig(vmCfgFile)
		if err != nil {
			return ret, errors.Wrap(err, "parsing VM config")
		}
		ret = append(ret, vmCfg)
	}
	return ret, nil
}

// GetVM will return a VMConfig for a VM identified by ID
func GetVM(vmID string) (VMConfig, error) {
	if vmID == "" {
		return VMConfig{}, fmt.Errorf("empty vmID")
	}
	allVms, err := ListAllVMs()
	if err != nil {
		return VMConfig{}, errors.Wrap(err, "fetching VM list")
	}

	for _, item := range allVms {
		if item.Name == vmID {
			return item, nil
		}
	}
	return VMConfig{}, fmt.Errorf("could not find VM with ID %s", vmID)
}
