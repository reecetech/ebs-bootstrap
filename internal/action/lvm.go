package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type CreatePhysicalVolumeAction struct {
	name       string
	mode       model.Mode
	lvmService service.LvmService
}

func NewCreatePhysicalVolumeAction(name string, ls service.LvmService) *CreatePhysicalVolumeAction {
	return &CreatePhysicalVolumeAction{
		name:       name,
		mode:       model.Empty,
		lvmService: ls,
	}
}

func (a *CreatePhysicalVolumeAction) Execute() error {
	return a.lvmService.CreatePhysicalVolume(a.name)
}

func (a *CreatePhysicalVolumeAction) GetMode() model.Mode {
	return a.mode
}

func (a *CreatePhysicalVolumeAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *CreatePhysicalVolumeAction) Prompt() string {
	return fmt.Sprintf("Would you like to create physical volume %s", a.name)
}

func (a *CreatePhysicalVolumeAction) Refuse() string {
	return fmt.Sprintf("Refused to create physical volume %s", a.name)
}

func (a *CreatePhysicalVolumeAction) Success() string {
	return fmt.Sprintf("Successfully created physical volume %s", a.name)
}
