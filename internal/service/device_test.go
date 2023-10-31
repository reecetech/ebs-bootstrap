package service

import (
	"testing"
	"fmt"
	"ebs-bootstrap/internal/utils"
	"github.com/google/go-cmp/cmp"
)

type deviceMockRunner struct {
	Output	string
	Error	error
}

func (mr *deviceMockRunner) Command(name string, arg ...string) (string, error) {
	return mr.Output, mr.Error
}

func TestGetBlockDevices(t *testing.T) {
	mr := &deviceMockRunner{
		Output: `{"blockdevices": [
					{"name":"nvme1n1", "label":"external-vol", "fstype":"xfs", "mountpoint":"/ifmx/dev/root"},
					{"name":"nvme0n1", "label":null, "fstype":null, "mountpoint":null}
				]}`,
		Error: nil,
	}
	expectedOutput := []string{"/dev/nvme1n1", "/dev/nvme0n1"}

	t.Run("Get Block Devices", func(t *testing.T) {
		du := &LinuxDeviceService{mr}
		d, err := du.GetBlockDevices()
		if !cmp.Equal(d, expectedOutput) {
			t.Errorf("GetBlockDevices() [output] mismatch: Expected=%v Actual=%v", expectedOutput, d)
		}
		utils.CheckError("GetBlockDevices()", t, nil, err)
	})
}

func TestGetDeviceInfo(t *testing.T) {
	deviceNotFoundErr := fmt.Errorf("🔴 /dev/nvme0n1 not a block device")

    subtests := []struct {
		Name		string
        Device      string
		MockRunner	*deviceMockRunner
		ExpectedOutput	*DeviceInfo
		ExpectedErr		error
    }{
        {
			Name:	"Get Device Info for /dev/nvme0n1",
            Device: "/dev/nvme0n1",
			MockRunner:	&deviceMockRunner{
				Output: `{"blockdevices":[{"name":"nvme0n1","label":"external-vol","fstype":"xfs","mountpoint":"/mnt/app"}]}`,
				Error: nil,
			},
			ExpectedOutput: &DeviceInfo{
				Name: "/dev/nvme0n1",
				Label: "external-vol",
				Fs: "xfs",
				MountPoint: "/mnt/app",
			},
			ExpectedErr: nil,
        },
        {
			Name:	"Get Device Info for /dev/nvme0n1 (No Fs,Label,Mountpoint)",
            Device: "/dev/nvme0n1",
			MockRunner:	&deviceMockRunner{
				Output: `{"blockdevices":[{"name":"nvme0n1","label":null,"fstype":null,"mountpoint":null}]}`,
				Error: nil,
			},
			ExpectedOutput: &DeviceInfo{
				Name: "/dev/nvme0n1",
				Label: "",
				Fs: "",
				MountPoint: "",
			},
			ExpectedErr: nil,
        },
        {
			Name:	"Get Device Info for Missing Device",
            Device: "/dev/nvme0n1",
			MockRunner:	&deviceMockRunner{
				Output: "",
				Error: deviceNotFoundErr,
			},
			ExpectedOutput: nil,
			ExpectedErr: deviceNotFoundErr,
        }, 
    }
    for _, subtest := range subtests {
        t.Run(subtest.Name, func(t *testing.T) {
			du := &LinuxDeviceService{subtest.MockRunner}
			di, err := du.GetDeviceInfo(subtest.Device)
			if !cmp.Equal(di, subtest.ExpectedOutput) {
				t.Errorf("GetDeviceInfo() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, di)
			}
            utils.CheckError("GetDeviceInfo()", t, subtest.ExpectedErr, err)
        })
    }
}

type mockDeviceService struct {
	getBlockDevices func() ([]string, error)
}

func (ds *mockDeviceService) GetBlockDevices() ([]string, error) {
	return ds.getBlockDevices()
}

func (ds *mockDeviceService) GetDeviceInfo(device string) (*DeviceInfo, error) {
	return nil, fmt.Errorf("🔴 GetDeviceInfo() not implemented")
}

type mockNVMeService struct {
	getBlockDeviceMapping func(device string)	(string, error)
}

func (ns *mockNVMeService) GetBlockDeviceMapping(device string)	(string, error) {
	return ns.getBlockDeviceMapping(device)
}

func TestEbsDeviceTranslator(t *testing.T) {
	subtests := []struct{
		Name	string
		DeviceService	DeviceService
		NVMeService		NVMeService
		ExpectedOutput	*DeviceTranslator
		ExpectedErr		error
	}{
		{
			Name: "Get DeviceTranslator for EBS NVME Device",
			DeviceService: &mockDeviceService{
				getBlockDevices: func() ([]string, error) {
					return []string{"/dev/nvme0n1"}, nil
				},
			},
			NVMeService: &mockNVMeService {
				getBlockDeviceMapping: func(device string)  (string, error) {
					return "/dev/xvdf", nil
				},
			},
			ExpectedOutput: &DeviceTranslator{
				Table: map[string]string{
					"/dev/nvme0n1" : "/dev/xvdf",
					"/dev/xvdf": "/dev/nvme0n1",
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Get DeviceTranslator for Traditional EBS Device",
			DeviceService: &mockDeviceService{
				getBlockDevices: func() ([]string, error) {
					return []string{"/dev/xvdf"}, nil
				},
			},
			NVMeService: &mockNVMeService{
				getBlockDeviceMapping: func(device string)  (string, error) {
					return "", fmt.Errorf("🔴 getBlockDeviceMapping() should not be called")
				},
			},
			ExpectedOutput: &DeviceTranslator{
				Table: map[string]string{
					"/dev/xvdf": "/dev/xvdf",
				},
			},
			ExpectedErr: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			dts := &EbsDeviceTranslator{
				DeviceService: subtest.DeviceService,
				NVMeService: subtest.NVMeService,
			}
			dt, err := dts.GetTranslator()
			if !cmp.Equal(dt, subtest.ExpectedOutput) {
				t.Errorf("GetTranslator() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, dt)
			}
            utils.CheckError("GetTranslator()", t, subtest.ExpectedErr, err)
		})
	}
}
