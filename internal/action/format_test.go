package action

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestFormatDeviceActionExecute(t *testing.T) {
	mfs := service.NewMockFileSystemService()
	mfs.StubFormat = func(device string) error { return nil }
	fda := NewFormatDeviceAction("/dev/xvdf", mfs)
	utils.ExpectErr("fda.Execute()", t, false, fda.Execute())
}

func TestFormatDeviceActionMode(t *testing.T) {
	fda := NewFormatDeviceAction("/dev/xvdf", nil)
	fda.SetMode(model.Healthcheck)
	utils.CheckOutput("fda.GetMode()", t, model.Healthcheck, fda.GetMode())
}

func TestFormatDeviceActionMessages(t *testing.T) {
	fda := NewFormatDeviceAction("/dev/xvdf", service.NewExt4Service(nil))
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        fda.Prompt(),
			ExpectedOutput: "Would you like to format /dev/xvdf to ext4",
		},
		{
			Name:           "Refuse",
			Message:        fda.Refuse(),
			ExpectedOutput: "Refused to format /dev/xvdf to ext4",
		},
		{
			Name:           "Success",
			Message:        fda.Success(),
			ExpectedOutput: "Successfully formatted /dev/xvdf to ext4",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}
