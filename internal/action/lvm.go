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

type CreateLogicalVolumeAction struct {
	name             string
	freeSpacePercent int
	volumeGroup      string
	mode             model.Mode
	lvmService       service.LvmService
}

func NewCreateLogicalVolumeAction(name string, freeSpacePercent int, volumeGroup string, ls service.LvmService) *CreateLogicalVolumeAction {
	return &CreateLogicalVolumeAction{
		name:             name,
		freeSpacePercent: freeSpacePercent,
		volumeGroup:      volumeGroup,
		mode:             model.Empty,
		lvmService:       ls,
	}
}

func (a *CreateLogicalVolumeAction) Execute() error {
	return a.lvmService.CreateLogicalVolume(a.name, a.volumeGroup, a.freeSpacePercent)
}

func (a *CreateLogicalVolumeAction) GetMode() model.Mode {
	return a.mode
}

func (a *CreateLogicalVolumeAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *CreateLogicalVolumeAction) Prompt() string {
	return fmt.Sprintf("Would you like to create logical volume %s that consumes %d%% free space of volume group %s", a.name, a.freeSpacePercent, a.volumeGroup)
}

func (a *CreateLogicalVolumeAction) Refuse() string {
	return fmt.Sprintf("Refused to create logical volume %s that consumes %d%% free space of volume group %s", a.name, a.freeSpacePercent, a.volumeGroup)
}

func (a *CreateLogicalVolumeAction) Success() string {
	return fmt.Sprintf("Successfully created logical volume %s that consumes %d%% free space of volume group %s", a.name, a.freeSpacePercent, a.volumeGroup)
}

type ActivateLogicalVolumeAction struct {
	name        string
	volumeGroup string
	mode        model.Mode
	lvmService  service.LvmService
}

func NewActivateLogicalVolumeAction(name string, volumeGroup string, ls service.LvmService) *ActivateLogicalVolumeAction {
	return &ActivateLogicalVolumeAction{
		name:        name,
		volumeGroup: volumeGroup,
		mode:        model.Empty,
		lvmService:  ls,
	}
}

func (a *ActivateLogicalVolumeAction) Execute() error {
	return a.lvmService.ActivateLogicalVolume(a.name, a.volumeGroup)
}

func (a *ActivateLogicalVolumeAction) GetMode() model.Mode {
	return a.mode
}

func (a *ActivateLogicalVolumeAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *ActivateLogicalVolumeAction) Prompt() string {
	return fmt.Sprintf("Would you like to activate logical volume %s in volume group %s", a.name, a.volumeGroup)
}

func (a *ActivateLogicalVolumeAction) Refuse() string {
	return fmt.Sprintf("Refused to activate logical volume %s in volume group %s", a.name, a.volumeGroup)
}

func (a *ActivateLogicalVolumeAction) Success() string {
	return fmt.Sprintf("Successfully activated logical volume %s in volume group %s", a.name, a.volumeGroup)
}
