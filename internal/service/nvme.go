package service

import (
    "fmt"
    "os"
    "syscall"
    "unsafe"
    "strings"
    "unicode"
)

const (
    NVME_ADMIN_IDENTIFY  = 0x06
    NVME_IOCTL_ADMIN_CMD = 0xC0484E41
    AMZN_NVME_VID        = 0x1D0F
    AMZN_NVME_EBS_MN     = "Amazon Elastic Block Store"
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
    Vid      uint16
    Ssvid    uint16
    Sn       [20]byte
    Mn       [40]byte
    Fr       [8]byte
    Rab      uint8
    Ieee     [3]uint8
    Mic      uint8
    Mdts     uint8
    Reserved0 [256 - 78]byte
    Oacs     uint16
    Acl      uint8
    Aerl     uint8
    Frmw     uint8
    Lpa      uint8
    Elpe     uint8
    Npss     uint8
    Avscc    uint8
    Reserved1 [512 - 265]byte
    Sqes     uint8
    Cqes     uint8
    Reserved2 uint16
    Nn       uint32
    Oncs     uint16
    Fuses    uint16
    Fna      uint8
    Vwc      uint8
    Awun     uint16
    Awupf    uint16
    Nvscc    uint8
    Reserved3 [704 - 531]byte
    Reserved4 [2048 - 704]byte
    Psd      [32]nvmeIdentifyControllerPSD
    Vs       nvmeIdentifyControllerAmznVS
}

type NVMeDevice struct {
    Name string
    IdCtrl nvmeIdentifyController
}

func NewNVMeDevice(name string) (*NVMeDevice, error) {
    d := &NVMeDevice{Name: name}
    if err := d.nvmeIOctl(); err != nil {
        return nil, err
    }
    return d, nil
}

func (d *NVMeDevice) nvmeIOctl() error {
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

// NVMe Service [Start]

type NVMeService interface {
	GetBlockDeviceMapping(device string)	(string, error)
}

// NVMe Service [END]

type AwsNVMeService struct {}

func (ns *AwsNVMeService) GetBlockDeviceMapping(device string) (string, error) {
	nd, err := NewNVMeDevice(device); 
    if err != nil {
        return "", err
    }
    return ns.getBlockDeviceMapping(nd)
}

func (ns *AwsNVMeService) isEBSVolume(nd *NVMeDevice) bool {
    vid := nd.IdCtrl.Vid
    mn := strings.TrimRightFunc(string(nd.IdCtrl.Mn[:]), unicode.IsSpace)
    return vid == AMZN_NVME_VID && mn == AMZN_NVME_EBS_MN
}

func (ns *AwsNVMeService) getBlockDeviceMapping(nd *NVMeDevice) (string, error) {
    var bdm string;
    if ns.isEBSVolume(nd) {
        bdm = strings.TrimRightFunc(string(nd.IdCtrl.Vs.Bdev[:]), unicode.IsSpace)
    }
    if bdm == "" {
        return "", fmt.Errorf("🔴 %s is not an AWS-managed NVME device", nd.Name)
    }
    if !strings.HasPrefix(bdm, "/dev/") {
        bdm = "/dev/" + bdm
    }
    return bdm, nil
}
