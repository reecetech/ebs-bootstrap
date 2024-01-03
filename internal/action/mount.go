package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type MountDeviceAction struct {
	source        string
	target        string
	fileSystem    model.FileSystem
	options       model.MountOptions
	deviceService service.DeviceService
	mode          model.Mode
}

func NewMountDeviceAction(source string, target string, fileSystem model.FileSystem, options model.MountOptions, deviceService service.DeviceService) *MountDeviceAction {
	return &MountDeviceAction{
		source:        source,
		target:        target,
		fileSystem:    fileSystem,
		options:       options,
		deviceService: deviceService,
		mode:          model.Empty,
	}
}

func (a *MountDeviceAction) Execute() error {
	return a.deviceService.Mount(a.source, a.target, a.fileSystem, a.options)
}

func (a *MountDeviceAction) GetMode() model.Mode {
	return a.mode
}

func (a *MountDeviceAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *MountDeviceAction) Prompt() string {
	return fmt.Sprintf("Would you like to mount %s to %s (%s)", a.source, a.target, a.options)
}

func (a *MountDeviceAction) Refuse() string {
	return fmt.Sprintf("Refused to mount %s to %s (%s)", a.source, a.target, a.options)
}

func (a *MountDeviceAction) Success() string {
	return fmt.Sprintf("Successfully mounted %s to %s (%s)", a.source, a.target, a.options)
}

type UnmountDeviceAction struct {
	source        string
	target        string
	deviceService service.DeviceService
	mode          model.Mode
}

func NewUnmountDeviceAction(source string, target string, deviceService service.DeviceService) *UnmountDeviceAction {
	return &UnmountDeviceAction{
		source:        source,
		target:        target,
		deviceService: deviceService,
		mode:          model.Empty,
	}
}

func (a *UnmountDeviceAction) Execute() error {
	return a.deviceService.Umount(a.source, a.target)
}

func (a *UnmountDeviceAction) GetMode() model.Mode {
	return a.mode
}

func (a *UnmountDeviceAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *UnmountDeviceAction) Prompt() string {
	return fmt.Sprintf("Would you like to unmount %s from %s", a.source, a.target)
}

func (a *UnmountDeviceAction) Refuse() string {
	return fmt.Sprintf("Refused to unmount %s from %s", a.source, a.target)
}

func (a *UnmountDeviceAction) Success() string {
	return fmt.Sprintf("Successfully unmounted %s from %s", a.source, a.target)
}
