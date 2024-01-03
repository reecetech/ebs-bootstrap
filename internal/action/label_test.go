package action

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestLabelDeviceActionExecute(t *testing.T) {
	mfs := service.NewMockFileSystemService()
	mfs.StubLabel = func(device string, label string) error { return nil }
	lda := NewLabelDeviceAction("/dev/xvdf", "example", mfs)
	utils.ExpectErr("lda.Execute()", t, false, lda.Execute())
}

func TestLabelDeviceActionMode(t *testing.T) {
	lda := NewLabelDeviceAction("/dev/xvdf", "example", nil)
	lda.SetMode(model.Healthcheck)
	utils.CheckOutput("lda.GetMode()", t, model.Healthcheck, lda.GetMode())
}

func TestLabelDeviceActionMessages(t *testing.T) {
	lda := NewLabelDeviceAction("/dev/xvdf", "example", nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        lda.Prompt(),
			ExpectedOutput: "Would you like to label device /dev/xvdf to 'example'",
		},
		{
			Name:           "Refuse",
			Message:        lda.Refuse(),
			ExpectedOutput: "Refused to label to 'example'",
		},
		{
			Name:           "Success",
			Message:        lda.Success(),
			ExpectedOutput: "Successfully labelled to 'example'",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}
