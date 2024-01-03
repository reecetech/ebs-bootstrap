package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type ResizeDeviceAction struct {
	device            string
	target            string
	fileSystemService service.FileSystemService
	mode              model.Mode
}

func NewResizeDeviceAction(d string, target string, fileSystemService service.FileSystemService) *ResizeDeviceAction {
	return &ResizeDeviceAction{
		device:            d,
		target:            target,
		fileSystemService: fileSystemService,
		mode:              model.Empty,
	}
}

func (a *ResizeDeviceAction) Execute() error {
	return a.fileSystemService.Resize(a.target)
}

func (a *ResizeDeviceAction) GetMode() model.Mode {
	return a.mode
}

func (a *ResizeDeviceAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *ResizeDeviceAction) Prompt() string {
	return fmt.Sprintf("Would you like to resize the %s file system of %s", a.fileSystemService.GetFileSystem(), a.device)
}

func (a *ResizeDeviceAction) Refuse() string {
	return fmt.Sprintf("Refused to resize the %s file system of %s", a.fileSystemService.GetFileSystem(), a.device)
}

func (a *ResizeDeviceAction) Success() string {
	return fmt.Sprintf("Successfully resized the %s file system of %s", a.fileSystemService.GetFileSystem(), a.device)
}
