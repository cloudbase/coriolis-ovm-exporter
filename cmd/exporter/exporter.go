package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"coriolis-ovm-exporter/internal"
	"coriolis-ovm-exporter/storage"
)

const (
	// FeCount is the default extent count we request
	FeCount = 8000
)

var (
	filename = flag.String("filepath", "", "path to file")
	dest     = flag.String("dest", "", "path to destination")

	vmID = flag.String("vmid", "", "VM ID to snapshot")
)

type deviceInfo struct {
	StartLBA   int
	BlockBytes int64
}

func findDev(major, minor uint64) (deviceInfo, error) {
	devs, err := storage.BlockDeviceList(false)
	if err != nil {
		fmt.Printf("Error listing block devices: %q\n", err)
		return deviceInfo{}, err
	}

	for _, device := range devs {
		if device.Major == major && device.Minor == minor {
			return deviceInfo{
				StartLBA:   0,
				BlockBytes: device.PhysicalSectorSize,
			}, nil
		} else if device.Partitions != nil {
			for _, part := range device.Partitions {
				if part.Major == major && part.Minor == minor {
					return deviceInfo{
						StartLBA:   part.StartSector,
						BlockBytes: device.PhysicalSectorSize,
					}, nil
				}
			}
		}
	}
	return deviceInfo{}, fmt.Errorf("device not found")
}

// func fibmapTest() {
// 	flag.Parse()

// 	if *filename == "" {
// 		flag.PrintDefaults()
// 		return
// 	}

// 	ext, err := internal.GetExtents(*filename)
// 	if err != nil {
// 		fmt.Printf("Error getting extents: %q\n", err)
// 		return
// 	}

// 	ee, err2 := json.MarshalIndent(ext, "", "  ")
// 	fmt.Println(string(ee), err2)
// 	fmt.Printf("total extents: %d\n", len(ext))
// }

func reflinkTest() {
	flag.Parse()

	if *filename == "" || *dest == "" {
		flag.PrintDefaults()
		return
	}
	if err := internal.IOctlOCFS2Reflink(*filename, *dest); err != nil {
		fmt.Printf("Error creating reflink: %q\n", err)
		return
	}
}

func main() {
	// reflinkTest()
	flag.Parse()

	if *vmID == "" {
		flag.PrintDefaults()
		return
	}

	vm, err := internal.GetVM(*vmID)
	if err != nil {
		fmt.Println(err)
		return
	}

	snap, err := vm.CreateSnapshot(false)
	if err != nil {
		fmt.Println(err)
		return
	}
	mm, err := json.MarshalIndent(snap, "", "  ")
	fmt.Println(string(mm), err)

	// cfg := config.Config{
	// 	APIServer: config.APIServer{
	// 		Bind: "0.0.0.0",
	// 		Port: 5544,
	// 		TLSConfig: config.TLSConfig{
	// 			CACert: "/home/gabriel/keys/ca-pub.pem",
	// 			Cert:   "/home/gabriel/keys/srv-pub.pem",
	// 			Key:    "/home/gabriel/keys/srv-key.pem",
	// 		},
	// 	},
	// }

	// repo := config.Repo{
	// 	Name:     "bogus",
	// 	FStype:   "ocfs",
	// 	Location: "/mnt",
	// }

	// repo2 := config.Repo{
	// 	Name:     "bogus2",
	// 	FStype:   "nfs",
	// 	Location: "/underwear",
	// }

	// cfg.Repos = append(cfg.Repos, repo)
	// cfg.Repos = append(cfg.Repos, repo2)

	// if err := cfg.Dump("/tmp/demo.toml"); err != nil {
	// 	fmt.Println(err)
	// }

	// vms, err := internal.ListAllVMs()

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// for _, vm := range vms {
	// 	disks, err := vm.Disks()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}

	// 	dd, err := json.MarshalIndent(disks, "", "  ")
	// 	fmt.Println(string(dd), err)

	// 	for _, disk := range disks {
	// 		fmt.Println(disk.CanClone())
	// 	}
	// }

	// vv, err := json.MarshalIndent(vms, "", "  ")
	// fmt.Println(string(vv), err)

	// repos, err := internal.ParseRepos()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// for _, val := range repos {
	// 	meta, err := val.Meta()
	// 	fmt.Println(meta, err)
	// }

	// mm, err := json.MarshalIndent(repos, "", "  ")
	// fmt.Println(string(mm), err)
	// fibmapTest()

	// finfo, err := fd.Stat()
	// if err != nil {
	// 	fmt.Printf("Failed to stat file: %q\n", err)
	// 	return
	// }
	// stat := finfo.Sys()

	// asStatT := stat.(*syscall.Stat_t)
	// fmt.Println(asStatT.Dev)
	// fmt.Printf("Device: %d:%d\n", storage.DeviceMajor(asStatT.Dev), storage.DeviceMinor(asStatT.Dev))

	// major := storage.DeviceMajor(asStatT.Dev)
	// minor := storage.DeviceMinor(asStatT.Dev)

	// devs, err := storage.BlockDeviceList(false)
	// if err != nil {
	// 	fmt.Printf("Error listing block devices: %q\n", err)
	// 	return
	// }
	// enc, err := json.MarshalIndent(devs, "", "  ")
	// fmt.Println(string(enc), err)

	// for _, device := range devs {
	// 	if device.Major == major && device.Minor == minor {
	// 		fmt.Printf("Device for %s is %s\n", *filename, device.Path)
	// 		break
	// 	} else if device.Partitions != nil {
	// 		for _, part := range device.Partitions {
	// 			if part.Major == major && part.Minor == minor {
	// 				fmt.Printf("Device for %s is %s\n", *filename, part.Path)
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	// blkSize, err := fmFile.Figetbsz()
	// if int(err.(syscall.Errno)) != 0 {
	// 	fmt.Printf("Failed to get blksize: %v\n", err)
	// 	return
	// }

	// fmt.Printf("BlockSize is %d\n", blkSize)

	// device, err := findDev(major, minor)
	// if err != nil {
	// 	fmt.Printf("Failed to find device with major %d amd minor %d\n", major, minor)
	// 	return
	// }
	// sectorsPerBlock := blkSize / int(device.BlockBytes)

	// fmt.Printf("Device is %v\n", device)

	// fmt.Println(fibmap.SEEK_SET)
	// fmt.Println("hi")
}
