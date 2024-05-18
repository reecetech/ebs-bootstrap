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

type CreateVolumeGroupAction struct {
	name           string
	physicalVolume string
	mode           model.Mode
	lvmService     service.LvmService
}

func NewCreateVolumeGroupAction(name string, physicalVolume string, ls service.LvmService) *CreateVolumeGroupAction {
	return &CreateVolumeGroupAction{
		name:           name,
		physicalVolume: physicalVolume,
		mode:           model.Empty,
		lvmService:     ls,
	}
}

func (a *CreateVolumeGroupAction) Execute() error {
	return a.lvmService.CreateVolumeGroup(a.name, a.physicalVolume)
}

func (a *CreateVolumeGroupAction) GetMode() model.Mode {
	return a.mode
}

func (a *CreateVolumeGroupAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *CreateVolumeGroupAction) Prompt() string {
	return fmt.Sprintf("Would you like to create volume group %s on physical volume %s", a.name, a.physicalVolume)
}

func (a *CreateVolumeGroupAction) Refuse() string {
	return fmt.Sprintf("Refused to create volume group %s on physical volume %s", a.name, a.physicalVolume)
}

func (a *CreateVolumeGroupAction) Success() string {
	return fmt.Sprintf("Successfully created volume group %s on physical volume %s", a.name, a.physicalVolume)
}
