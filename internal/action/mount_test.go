package action

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestMountDeviceActionExecute(t *testing.T) {
	mds := service.NewMockDeviceService()
	mds.StubMount = func(source string, target string, fs model.FileSystem, options model.MountOptions) error { return nil }
	mda := NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "defaults", mds)
	utils.ExpectErr("mda.Execute()", t, false, mda.Execute())
}

func TestMountDeviceActionMode(t *testing.T) {
	mda := NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "defaults", nil)
	mda.SetMode(model.Healthcheck)
	utils.CheckOutput("mda.GetMode()", t, model.Healthcheck, mda.GetMode())
}

func TestMountDeviceActionMessages(t *testing.T) {
	mda := NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "defaults", nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        mda.Prompt(),
			ExpectedOutput: "Would you like to mount /dev/xvdf to /mnt/foo (defaults)",
		},
		{
			Name:           "Refuse",
			Message:        mda.Refuse(),
			ExpectedOutput: "Refused to mount /dev/xvdf to /mnt/foo (defaults)",
		},
		{
			Name:           "Success",
			Message:        mda.Success(),
			ExpectedOutput: "Successfully mounted /dev/xvdf to /mnt/foo (defaults)",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}

func TestUnmountDeviceActionExecute(t *testing.T) {
	mds := service.NewMockDeviceService()
	mds.StubUmount = func(source string, target string) error { return nil }
	uda := NewUnmountDeviceAction("/dev/xvdf", "/mnt/foo", mds)
	utils.ExpectErr("uda.Execute()", t, false, uda.Execute())
}

func TestUnmountDeviceActionMode(t *testing.T) {
	uda := NewUnmountDeviceAction("/dev/xvdf", "/mnt/foo", nil)
	uda.SetMode(model.Healthcheck)
	utils.CheckOutput("uda.GetMode()", t, model.Healthcheck, uda.GetMode())
}

func TestUnmountDeviceActionMessages(t *testing.T) {
	uda := NewUnmountDeviceAction("/dev/xvdf", "/mnt/foo", nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        uda.Prompt(),
			ExpectedOutput: "Would you like to unmount /dev/xvdf from /mnt/foo",
		},
		{
			Name:           "Refuse",
			Message:        uda.Refuse(),
			ExpectedOutput: "Refused to unmount /dev/xvdf from /mnt/foo",
		},
		{
			Name:           "Success",
			Message:        uda.Success(),
			ExpectedOutput: "Successfully unmounted /dev/xvdf from /mnt/foo",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}
