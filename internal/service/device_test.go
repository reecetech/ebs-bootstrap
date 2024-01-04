package service

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetSize(t *testing.T) {
	subtests := []struct {
		Name           string
		Device         string
		RunnerBinary   utils.Binary
		RunnerArgs     []string
		RunnerOutput   string
		RunnerError    error
		ExpectedOutput uint64
		ExpectedError  error
	}{
		{
			Name:           "blockdev=success + cast=success",
			Device:         "/dev/vdb",
			RunnerBinary:   utils.BlockDev,
			RunnerArgs:     []string{"--getsize64", "/dev/vdb"},
			RunnerOutput:   "12345",
			RunnerError:    nil,
			ExpectedOutput: 12345,
			ExpectedError:  nil,
		},
		{
			Name:           "blockdev=success + cast=failure",
			Device:         "/dev/vdc",
			RunnerBinary:   utils.BlockDev,
			RunnerArgs:     []string{"--getsize64", "/dev/vdc"},
			RunnerOutput:   "lsblk: permission denied",
			RunnerError:    nil,
			ExpectedOutput: 0,
			ExpectedError:  fmt.Errorf("🔴 Failed to cast block device size to unsigned 64-bit integer"),
		},
		{
			Name:           "blockdev=error",
			Device:         "/dev/vdd",
			RunnerBinary:   utils.BlockDev,
			RunnerArgs:     []string{"--getsize64", "/dev/vdd"},
			RunnerOutput:   "",
			RunnerError:    fmt.Errorf("🔴 blockdev is either not installed or accessible from $PATH"),
			ExpectedOutput: 0,
			ExpectedError:  fmt.Errorf("🔴 blockdev is either not installed or accessible from $PATH"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			lds := NewLinuxDeviceService(mrf)
			size, err := lds.GetSize(subtest.Device)
			utils.CheckError("lds.GetSize()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lds.GetSize()", t, subtest.ExpectedOutput, size)
		})
	}
}

func TestGetBlockDevices(t *testing.T) {
	subtests := []struct {
		Name           string
		RunnerBinary   utils.Binary
		RunnerArgs     []string
		RunnerOutput   string
		RunnerError    error
		ExpectedOutput []string
		ExpectedError  error
	}{
		{
			Name:         "lsblk=success + json=valid",
			RunnerBinary: utils.Lsblk,
			RunnerArgs:   []string{"--nodeps", "-o", "NAME", "-J"},
			RunnerOutput: `{"blockdevices": [
				{"name":"nvme1n1"},
				{"name":"nvme0n1"}
			]}`,
			RunnerError:    nil,
			ExpectedOutput: []string{"/dev/nvme1n1", "/dev/nvme0n1"},
			ExpectedError:  nil,
		},
		{
			Name:           "lsblk=success + json=invalid",
			RunnerBinary:   utils.Lsblk,
			RunnerArgs:     []string{"--nodeps", "-o", "NAME", "-J"},
			RunnerOutput:   `{"invalid_json"}`,
			RunnerError:    nil,
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 Failed to decode lsblk response: *"),
		},
		{
			Name:           "lsblk=failure",
			RunnerBinary:   utils.Lsblk,
			RunnerArgs:     []string{"--nodeps", "-o", "NAME", "-J"},
			RunnerOutput:   "",
			RunnerError:    fmt.Errorf("🔴 lsblk: invalid option -- 'J'"),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 lsblk: invalid option -- 'J'"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			lds := NewLinuxDeviceService(mrf)
			size, err := lds.GetBlockDevices()
			utils.CheckErrorGlob("lds.GetBlockDevices()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lds.GetBlockDevices()", t, subtest.ExpectedOutput, size)
		})
	}
}

func TestGetBlockDevice(t *testing.T) {
	subtests := []struct {
		Name           string
		Device         string
		RunnerBinary   utils.Binary
		RunnerArgs     []string
		RunnerOutput   string
		RunnerError    error
		ExpectedOutput *model.BlockDevice
		ExpectedError  error
	}{
		{
			Name:         "lsblk=success",
			Device:       "/dev/nvme1n1",
			RunnerBinary: utils.Lsblk,
			RunnerArgs:   []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/nvme1n1"},
			RunnerOutput: `{"blockdevices": [
				{"name":"nvme1n1", "label":"external-vol", "fstype":"xfs", "mountpoint":"/mnt/app"}
			]}`,
			RunnerError: nil,
			ExpectedOutput: &model.BlockDevice{
				Name:       "/dev/nvme1n1",
				Label:      "external-vol",
				FileSystem: model.Xfs,
				MountPoint: "/mnt/app",
			},
			ExpectedError: nil,
		},
		{
			Name:         "lsblk=success + device=unformatted",
			Device:       "/dev/nvme1n1",
			RunnerBinary: utils.Lsblk,
			RunnerArgs:   []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/nvme1n1"},
			RunnerOutput: `{"blockdevices": [
				{"name":"nvme1n1", "label":null, "fstype":null, "mountpoint":null}
			]}`,
			RunnerError: nil,
			ExpectedOutput: &model.BlockDevice{
				Name:       "/dev/nvme1n1",
				Label:      "",
				FileSystem: model.Unformatted,
				MountPoint: "",
			},
			ExpectedError: nil,
		},
		/*	Reference [https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html]
			---
			For example, if /dev/sdb is renamed /dev/xvdf, then /dev/sdc is renamed /dev/xvdg.
			Amazon Linux creates a symbolic link for the name you specified to the renamed device.
			Other operating systems could behave differently.
			--
			With this behaviour established on Amazon Linux, it is possible for lsblk to return a
			device name that might not reflect the one that was provided to lsblk. To simplify the
			output of ebs-bootstrap, lets force the BlockDevice to house the name that was provided
			to the lsblk call
		*/
		{
			Name:         "lsblk=success + symlink=true",
			Device:       "/dev/sdb",
			RunnerBinary: utils.Lsblk,
			RunnerArgs:   []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/sdb"},
			RunnerOutput: `{"blockdevices": [
				{"name":"xvdf", "label":"external-vol", "fstype":"xfs", "mountpoint":"/mnt/app"}
			]}`,
			RunnerError: nil,
			ExpectedOutput: &model.BlockDevice{
				Name:       "/dev/sdb",
				Label:      "external-vol",
				FileSystem: model.Xfs,
				MountPoint: "/mnt/app",
			},
			ExpectedError: nil,
		},
		{
			Name:           "lsblk=success + json=invalid",
			Device:         "/dev/sdb",
			RunnerBinary:   utils.Lsblk,
			RunnerArgs:     []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/sdb"},
			RunnerOutput:   `{"invalid_json"}`,
			RunnerError:    nil,
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 Failed to decode lsblk response: *"),
		},
		{
			Name:         "lsblk=success + filesystem=unsupported",
			Device:       "/dev/sdb",
			RunnerBinary: utils.Lsblk,
			RunnerArgs:   []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/sdb"},
			RunnerOutput: `{"blockdevices": [
				{"name":"sdb", "label":null, "fstype":"jfs", "mountpoint":"/mnt/app"}
			]}`,
			RunnerError:    nil,
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/sdb: File system 'jfs' is not supported"),
		},
		/*	I haven't encountered a scenario where lsblk successfully returns an empty array
			Typically, if it cannot find a specific block device, it would produce an error.
			Nevertheless, lets test this scenario...
		*/
		{
			Name:           "lsblk=success + len(devices)==0",
			Device:         "/dev/sdb",
			RunnerBinary:   utils.Lsblk,
			RunnerArgs:     []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/sdb"},
			RunnerOutput:   `{"blockdevices": []}`,
			RunnerError:    nil,
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/sdb: An unexpected number of block devices were returned: Expected=1 Actual=0"),
		},
		{
			Name:           "lsblk=failure",
			Device:         "/dev/sdc",
			RunnerBinary:   utils.Lsblk,
			RunnerArgs:     []string{"--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", "/dev/sdc"},
			RunnerOutput:   "",
			RunnerError:    fmt.Errorf("🔴 lsblk: /dev/sdc: not a block device"),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 lsblk: /dev/sdc: not a block device"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			lds := NewLinuxDeviceService(mrf)
			size, err := lds.GetBlockDevice(subtest.Device)
			utils.CheckErrorGlob("lds.GetBlockDevice()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lds.GetBlockDevice()", t, subtest.ExpectedOutput, size)
		})
	}
}

func TestMount(t *testing.T) {
	subtests := []struct {
		Name          string
		Source        string
		Target        string
		Fs            model.FileSystem
		Options       model.MountOptions
		RunnerBinary  utils.Binary
		RunnerArgs    []string
		RunnerOutput  string
		RunnerError   error
		ExpectedError error
	}{
		{
			Name:          "mount=success",
			Source:        "/dev/sdb",
			Target:        "/mnt/data",
			Fs:            model.Ext4,
			Options:       model.MountOptions("defaults"),
			RunnerBinary:  utils.Mount,
			RunnerArgs:    []string{"/dev/sdb", "-t", "ext4", "-o", "defaults", "/mnt/data"},
			RunnerOutput:  "",
			RunnerError:   nil,
			ExpectedError: nil,
		},
		{
			Name:          "mount=failure",
			Source:        "/dev/sdc",
			Target:        "/mnt/data",
			Fs:            model.Xfs,
			Options:       model.MountOptions("defaults"),
			RunnerBinary:  utils.Mount,
			RunnerArgs:    []string{"/dev/sdc", "-t", "xfs", "-o", "defaults", "/mnt/data"},
			RunnerOutput:  "",
			RunnerError:   fmt.Errorf("🔴 mount: /dev/sdc: special device /mnt/data does not exist"),
			ExpectedError: fmt.Errorf("🔴 mount: /dev/sdc: special device /mnt/data does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			lds := NewLinuxDeviceService(mrf)
			err := lds.Mount(subtest.Source, subtest.Target, subtest.Fs, subtest.Options)
			utils.CheckErrorGlob("lds.Mount()", t, subtest.ExpectedError, err)
		})
	}
}

func TestUmount(t *testing.T) {
	subtests := []struct {
		Name          string
		Source        string
		Target        string
		RunnerBinary  utils.Binary
		RunnerArgs    []string
		RunnerOutput  string
		RunnerError   error
		ExpectedError error
	}{
		{
			Name:          "umount=success",
			Source:        "/dev/sdb",
			Target:        "/mnt/data",
			RunnerBinary:  utils.Umount,
			RunnerArgs:    []string{"/mnt/data"},
			RunnerOutput:  "",
			RunnerError:   nil,
			ExpectedError: nil,
		},
		{
			Name:          "umount=failure",
			Source:        "/dev/sdc",
			Target:        "/mnt/data",
			RunnerBinary:  utils.Umount,
			RunnerArgs:    []string{"/mnt/data"},
			RunnerOutput:  "",
			RunnerError:   fmt.Errorf("🔴 umount: /mnt/data: not mounted"),
			ExpectedError: fmt.Errorf("🔴 umount: /mnt/data: not mounted"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			lds := NewLinuxDeviceService(mrf)
			err := lds.Umount(subtest.Source, subtest.Target)
			utils.CheckErrorGlob("lds.Umount()", t, subtest.ExpectedError, err)
		})
	}
}
