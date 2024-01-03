package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type FormatDeviceAction struct {
	device            string
	fileSystemService service.FileSystemService
	mode              model.Mode
}

func NewFormatDeviceAction(d string, fileSystemService service.FileSystemService) *FormatDeviceAction {
	return &FormatDeviceAction{
		device:            d,
		fileSystemService: fileSystemService,
		mode:              model.Empty,
	}
}

func (a *FormatDeviceAction) Execute() error {
	return a.fileSystemService.Format(a.device)
}

func (a *FormatDeviceAction) GetMode() model.Mode {
	return a.mode
}

func (a *FormatDeviceAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *FormatDeviceAction) Prompt() string {
	return fmt.Sprintf("Would you like to format %s to %s", a.device, a.fileSystemService.GetFileSystem())
}

func (a *FormatDeviceAction) Refuse() string {
	return fmt.Sprintf("Refused to format to %s", a.fileSystemService.GetFileSystem())
}

func (a *FormatDeviceAction) Success() string {
	return fmt.Sprintf("Successfully formatted to %s", a.fileSystemService.GetFileSystem().String())
}
