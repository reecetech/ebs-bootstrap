package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type LabelDeviceAction struct {
	device            string
	label             string
	fileSystemService service.FileSystemService
	mode              model.Mode
}

func NewLabelDeviceAction(d string, label string, fileSystemService service.FileSystemService) *LabelDeviceAction {
	return &LabelDeviceAction{
		device:            d,
		label:             label,
		fileSystemService: fileSystemService,
		mode:              model.Empty,
	}
}

func (a *LabelDeviceAction) Execute() error {
	return a.fileSystemService.Label(a.device, a.label)
}

func (a *LabelDeviceAction) GetMode() model.Mode {
	return a.mode
}

func (a *LabelDeviceAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *LabelDeviceAction) Prompt() string {
	return fmt.Sprintf("Would you like to label device %s to '%s'", a.device, a.label)
}

func (a *LabelDeviceAction) Refuse() string {
	return fmt.Sprintf("Refused to label to '%s'", a.label)
}

func (a *LabelDeviceAction) Success() string {
	return fmt.Sprintf("Successfully labelled to '%s'", a.label)
}
