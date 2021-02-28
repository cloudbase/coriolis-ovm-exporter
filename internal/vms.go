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
	"github.com/pkg/errors"
)

const (
	// VirtualMachinesDir is the directory where VM config files
	// are stored.
	VirtualMachinesDir = "VirtualMachines"
)

// Disk represents one VM disk
type Disk struct {
	Path       string
	DeviceName string
	Mode       string
}

// CanClone returns a boolean value indicating whether or not
// this disk can be reflinked.
func (d Disk) CanClone() bool {
	repo, err := d.Repo()
	if err != nil {
		log.Printf("Failed to get repo: %q", err)
		return false
	}

	if repo.Filesystem != "ocfs2" {
		log.Printf("Filesystem is not ocfs2: %s", repo.Filesystem)
		return false
	}

	return true
}

// Repo returns the repo where this disk is hosted
func (d Disk) Repo() (Repo, error) {
	repos, err := ParseRepos()
	if err != nil {
		return Repo{}, errors.Wrap(err, "fetching repos")
	}
	baseName := filepath.Base(d.Path)

	for _, repo := range repos {
		meta, err := repo.Meta()
		if err != nil {
			log.Printf("failed to find repo metadata: %q", err)
		} else {
			if diskMeta, ok := meta[baseName]; ok {
				if diskMeta.ObjectType == "VIRTUAL_DISK" {
					return repo, nil
				} else {
					// No need to attempt prefix match
					continue
				}
			}
		}

		// attempt a prefix match
		if strings.HasPrefix(d.Path, repo.MountPoint) {
			return repo, nil
		}
	}

	return Repo{}, fmt.Errorf("could not find repo for disk %s", d.Path)
}

// VMConfig is a stripped down VM config, containing only
// the fields we care about.
type VMConfig struct {
	// OVM_simple_name is the friendly name for a VM
	OVMSimpleName string `toml:"OVM_simple_name"`
	// Name is the internal name of the VM
	Name string
	// UUID is the UUID of the VM
	UUID string
	// Disk is a list of paths to virtual machine disks
	DiskArray []string `toml:"disk"`
}

// Disks returns an array of Disk objects, representing the
// disks attached to a VM.
func (v VMConfig) Disks() ([]Disk, error) {
	var ret []Disk
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

		ret = append(ret, Disk{
			Path:       details[0],
			DeviceName: details[1],
			Mode:       details[2],
		})
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
