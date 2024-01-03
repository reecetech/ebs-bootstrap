package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestAwsNitroNVMeModifier(t *testing.T) {
	subtests := []struct {
		Name                  string
		Config                *Config
		GetBlockDevices       func() ([]string, error)
		GetBlockDeviceMapping func(name string) (string, error)
		ExpectedOutput        *Config
		ExpectedError         error
	}{
		{
			Name: "Root Device + EBS Device (Non-Nitro Instance)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			GetBlockDevices: func() ([]string, error) {
				return []string{"/dev/sda1", "/dev/sdb"}, nil
			},
			ExpectedOutput: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Root Device + EBS/Instance Store Device (Nitro Instance)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			GetBlockDevices: func() ([]string, error) {
				return []string{"/dev/nvme0n1", "/dev/nvme1n1"}, nil
			},
			GetBlockDeviceMapping: func(name string) (string, error) {
				switch name {
				case "/dev/nvme0n1": // Root Device
					return "/dev/sda1", nil
				default: // EBS/Instance Store
					return "/dev/sdb", nil
				}
			},
			// Config will be left unchanged when error is encountered during modification stage
			ExpectedOutput: &Config{
				Devices: map[string]Device{
					"/dev/nvme1n1": {},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "NVMe Device that is not AWS-managed",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			GetBlockDevices: func() ([]string, error) {
				return []string{"/dev/nvme0n1"}, nil
			},
			GetBlockDeviceMapping: func(name string) (string, error) {
				return "", fmt.Errorf("🔴 %s is not an AWS-managed NVME device", name)
			},
			ExpectedOutput: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/nvme0n1 is not an AWS-managed NVME device"),
		},
		{
			Name: "Failure to Retrieve Block Devices",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			GetBlockDevices: func() ([]string, error) {
				return nil, fmt.Errorf("🔴 lsblk: Could not retrieve block devices")
			},
			// Config will be left unchanged when error is encountered during modification stage
			ExpectedOutput: &Config{
				Devices: map[string]Device{
					"/dev/sdb": {},
				},
			},
			ExpectedError: fmt.Errorf("🔴 lsblk: Could not retrieve block devices"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ds := service.NewMockDeviceService()
			if subtest.GetBlockDevices != nil {
				ds.StubGetBlockDevices = subtest.GetBlockDevices
			}
			ns := service.NewMockNVMeService()
			if subtest.GetBlockDeviceMapping != nil {
				ns.StubGetBlockDeviceMapping = subtest.GetBlockDeviceMapping
			}

			andm := NewAwsNVMeDriverModifier(ns, ds)
			err := andm.Modify(subtest.Config)
			utils.CheckError("andm.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("andm.Modify()", t, subtest.ExpectedOutput, subtest.Config, cmp.AllowUnexported(Config{}))
		})
	}
}
