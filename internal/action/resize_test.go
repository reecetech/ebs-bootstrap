package action

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestResizeDeviceActionExecute(t *testing.T) {
	mfs := service.NewMockFileSystemService()
	mfs.StubResize = func(name string) error { return nil }
	rda := NewResizeDeviceAction("/dev/xvdf", "/mnt/foo", mfs)
	utils.ExpectErr("rda.Execute()", t, false, rda.Execute())
}

func TestResizeDeviceActionMode(t *testing.T) {
	rda := NewResizeDeviceAction("/dev/xvdf", "/mnt/foo", nil)
	rda.SetMode(model.Healthcheck)
	utils.CheckOutput("rda.GetMode()", t, model.Healthcheck, rda.GetMode())
}

func TestResizeDeviceActionMessages(t *testing.T) {
	rda := NewResizeDeviceAction("/dev/xvdf", "/mnt/foo", service.NewExt4Service(nil))
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        rda.Prompt(),
			ExpectedOutput: "Would you like to resize the ext4 file system of /dev/xvdf",
		},
		{
			Name:           "Refuse",
			Message:        rda.Refuse(),
			ExpectedOutput: "Refused to resize the ext4 file system of /dev/xvdf",
		},
		{
			Name:           "Success",
			Message:        rda.Success(),
			ExpectedOutput: "Successfully resized the ext4 file system of /dev/xvdf",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}
