package service

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

const (
	NVME_ADMIN_IDENTIFY  = 0x06
	NVME_IOCTL_ADMIN_CMD = 0xC0484E41
	AMZN_NVME_VID        = 0x1D0F
	AMZN_NVME_EBS_MN     = "Amazon Elastic Block Store"
	AMZN_NVME_INS_MN     = "Amazon EC2 NVMe Instance Storage"
)

type nvmeAdminCommand struct {
	Opcode    uint8
	Flags     uint8
	Cid       uint16
	Nsid      uint32
	Reserved0 uint64
	Mptr      uint64
	Addr      uint64
	Mlen      uint32
	Alen      uint32
	Cdw10     uint32
	Cdw11     uint32
	Cdw12     uint32
	Cdw13     uint32
	Cdw14     uint32
	Cdw15     uint32
	Reserved1 uint64
}

type nvmeIdentifyControllerAmznVS struct {
	Bdev      [32]byte
	Reserved0 [1024 - 32]byte
}

type nvmeIdentifyControllerPSD struct {
	Mp        uint16
	Reserved0 uint16
	Enlat     uint32
	Exlat     uint32
	Rrt       uint8
	Rrl       uint8
	Rwt       uint8
	Rwl       uint8
	Reserved1 [16]byte
}

type nvmeIdentifyController struct {
	Vid       uint16
	Ssvid     uint16
	Sn        [20]byte
	Mn        [40]byte
	Fr        [8]byte
	Rab       uint8
	Ieee      [3]uint8
	Mic       uint8
	Mdts      uint8
	Reserved0 [256 - 78]byte
	Oacs      uint16
	Acl       uint8
	Aerl      uint8
	Frmw      uint8
	Lpa       uint8
	Elpe      uint8
	Npss      uint8
	Avscc     uint8
	Reserved1 [512 - 265]byte
	Sqes      uint8
	Cqes      uint8
	Reserved2 uint16
	Nn        uint32
	Oncs      uint16
	Fuses     uint16
	Fna       uint8
	Vwc       uint8
	Awun      uint16
	Awupf     uint16
	Nvscc     uint8
	Reserved3 [704 - 531]byte
	Reserved4 [2048 - 704]byte
	Psd       [32]nvmeIdentifyControllerPSD
	Vs        nvmeIdentifyControllerAmznVS
}

var instanceStoreRegex = regexp.MustCompile(`^(ephemeral[0-9]):(sd[a-z]|none)`)

type NVMeIoctlResult struct {
	Name   string
	IdCtrl nvmeIdentifyController
}

func NewNVMeIoctlResult(name string) *NVMeIoctlResult {
	return &NVMeIoctlResult{Name: name}
}

func (d *NVMeIoctlResult) Syscall() error {
	idResponse := uintptr(unsafe.Pointer(&d.IdCtrl))
	idLen := unsafe.Sizeof(d.IdCtrl)

	adminCmd := nvmeAdminCommand{
		Opcode: NVME_ADMIN_IDENTIFY,
		Addr:   uint64(idResponse),
		Alen:   uint32(idLen),
		Cdw10:  1,
	}

	nvmeFile, err := os.OpenFile(d.Name, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer nvmeFile.Close()

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, nvmeFile.Fd(), NVME_IOCTL_ADMIN_CMD, uintptr(unsafe.Pointer(&adminCmd)))
	if errno != 0 {
		return fmt.Errorf("🔴 ioctl error: %v", errno)
	}

	return nil
}

type NVMeService interface {
	GetBlockDeviceMapping(device string) (string, error)
}

type AwsNitroNVMeService struct{}

func NewAwsNitroNVMeService() *AwsNitroNVMeService {
	return &AwsNitroNVMeService{}
}

func (ns *AwsNitroNVMeService) GetBlockDeviceMapping(device string) (string, error) {
	nir := NewNVMeIoctlResult(device)
	if err := nir.Syscall(); err != nil {
		return "", err
	}
	return ns.getBlockDeviceMapping(nir)
}

func (ns *AwsNitroNVMeService) isEBSVolume(nir *NVMeIoctlResult) bool {
	vid := nir.IdCtrl.Vid
	mn := strings.TrimRightFunc(string(nir.IdCtrl.Mn[:]), ns.trimModelNumber)
	return vid == AMZN_NVME_VID && mn == AMZN_NVME_EBS_MN
}

func (ns *AwsNitroNVMeService) isInstanceStoreVolume(nir *NVMeIoctlResult) bool {
	vid := nir.IdCtrl.Vid
	mn := strings.TrimRightFunc(string(nir.IdCtrl.Mn[:]), ns.trimModelNumber)
	return vid == AMZN_NVME_VID && mn == AMZN_NVME_INS_MN
}

func (ns *AwsNitroNVMeService) getBlockDeviceMapping(nir *NVMeIoctlResult) (string, error) {
	var bdm string
	if ns.isEBSVolume(nir) {
		bdm = strings.TrimRightFunc(string(nir.IdCtrl.Vs.Bdev[:]), ns.trimBlockDevice)
	}
	if ns.isInstanceStoreVolume(nir) {
		// Vendor Specfic (vs)
		vs := strings.TrimRightFunc(string(nir.IdCtrl.Vs.Bdev[:]), ns.trimBlockDevice)

		// Match Block Device Mapping
		mbdm := instanceStoreRegex.FindStringSubmatch(vs)
		if len(mbdm) != 3 {
			return "", fmt.Errorf("🔴 %s: Instance-store vendor specific metadata did not match pattern. Pattern=%s, Actual=%s", nir.Name, instanceStoreRegex.String(), vs)
		}
		// If the block device mapping is "none", then lets default to assigning the
		// the block device mapping to the match result from ephemeral[0-9]
		if mbdm[2] == "none" {
			bdm = mbdm[1]
		} else {
			bdm = mbdm[2]
		}
	}
	if len(bdm) == 0 {
		return "", fmt.Errorf("🔴 %s is not an AWS-managed NVME device", nir.Name)
	}
	if !strings.HasPrefix(bdm, "/dev/") {
		bdm = "/dev/" + bdm
	}
	return bdm, nil
}

func (ns *AwsNitroNVMeService) trimModelNumber(r rune) bool {
	// Explanation:
	// 	- Both the AWS EC2 and EBS team use the 0x20 (space) byte to pad out the Model Number
	return r == 0x20
}

func (ns *AwsNitroNVMeService) trimBlockDevice(r rune) bool {
	// Explanation:
	// 	- The AWS EC2 team uses the 0x00 (null) byte to pad out the Vendor Specific allocation
	//	- The AWS EBS team uses the 0x20 (space) byte to pad out the Vendor Specific allocation
	//	- It is frustrating that the padding character is not standardised, but we can
	//	  work around this by checking for both bytes when trimming the Vendor Specific allocation
	// Examples:
	// 	nd.IdCtrl.Vs.Bdev[:] ("Amazon EC2 NVMe Instance Storage")
	// 		  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	// 0000: 65 70 68 65 6d 65 72 61 6c 30 3a 73 64 68 00 00 "ephemeral0:sdh.."
	// 0010: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 "................"
	//	nd.IdCtrl.Vs.Bdev[:] (Amazon Elastic Block Store)
	//        0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	// 0000: 2f 64 65 76 2f 73 64 63 20 20 20 20 20 20 20 20 "/dev/sdc........"
	// 0010: 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 "................"
	return r == 0x00 || r == 0x20
}
