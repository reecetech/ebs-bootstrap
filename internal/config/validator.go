package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type Validator interface {
	Validate(c *Config) error
}

type DeviceValidator struct {
	deviceService service.DeviceService
}

func NewDeviceValidator(ds service.DeviceService) *DeviceValidator {
	return &DeviceValidator{
		deviceService: ds,
	}
}

func (dv *DeviceValidator) Validate(c *Config) error {
	for name := range c.Devices {
		_, err := dv.deviceService.GetBlockDevice(name)
		if err != nil {
			return err
		}
	}
	return nil
}

type FileSystemValidator struct{}

func NewFileSystemValidator() *FileSystemValidator {
	return &FileSystemValidator{}
}

func (fsv *FileSystemValidator) Validate(c *Config) error {
	for name, device := range c.Devices {
		fs, err := model.ParseFileSystem(string(device.Fs))
		if err != nil {
			return fmt.Errorf("🔴 %s: %s", name, err)
		}
		if fs == model.Unformatted {
			return fmt.Errorf("🔴 %s: Must provide a supported file system", name)
		}
	}
	return nil
}

type ModeValidator struct{}

func NewModeValidator() *ModeValidator {
	return &ModeValidator{}
}

func (fsv *ModeValidator) Validate(c *Config) error {
	mode := string(c.Defaults.Mode)
	_, err := model.ParseMode(mode)
	if err != nil {
		return fmt.Errorf("🔴 '%s' (defaults) is not a supported mode", mode)
	}

	mode = string(c.overrides.Mode)
	_, err = model.ParseMode(mode)
	if err != nil {
		return fmt.Errorf("🔴 '%s' (-mode) is not a supported mode", mode)
	}

	for name, device := range c.Devices {
		mode := string(device.Mode)
		_, err := model.ParseMode(mode)
		if err != nil {
			return fmt.Errorf("🔴 %s: '%s' is not a supported mode", name, mode)
		}
	}
	return nil
}

type MountPointValidator struct{}

func NewMountPointValidator() *MountPointValidator {
	return &MountPointValidator{}
}

func (apv *MountPointValidator) Validate(c *Config) error {
	for name, device := range c.Devices {
		if len(device.MountPoint) > 0 {
			if !path.IsAbs(device.MountPoint) {
				return fmt.Errorf("🔴 %s: %s is not an absolute path", name, device.MountPoint)
			}
			if device.MountPoint == "/" {
				return fmt.Errorf("🔴 %s: Can not be mounted to the root directory", name)
			}
		}
	}
	return nil
}

type MountOptionsValidator struct{}

func NewMountOptionsValidator() *MountOptionsValidator {
	return &MountOptionsValidator{}
}

func (mov *MountOptionsValidator) Validate(c *Config) error {
	mo := string(c.Defaults.MountOptions)
	if err := mov.validate(mo); err != nil {
		return fmt.Errorf("🔴 '%s' (defaults) is not a supported mode as %s", mo, err)
	}
	mo = string(c.overrides.MountOptions)
	if err := mov.validate(mo); err != nil {
		return fmt.Errorf("🔴 '%s' (-mount-options) is not a supported mode as %s", mo, err)
	}
	for name, device := range c.Devices {
		mo := string(device.MountOptions)
		if err := mov.validate(mo); err != nil {
			return fmt.Errorf("🔴 %s: '%s' is not a supported mode as %s", name, mo, err)
		}
	}
	return nil
}

func (mov *MountOptionsValidator) validate(mo string) error {
	if strings.Contains(mo, "remount") {
		return fmt.Errorf("it prevents unmounted devices from being mounted")
	}
	if strings.Contains(mo, "bind") {
		return fmt.Errorf("bind mounts are not supported for block devices")
	}
	return nil
}

type OwnerValidator struct {
	ownerService service.OwnerService
}

func NewOwnerValidator(ows service.OwnerService) *OwnerValidator {
	return &OwnerValidator{
		ownerService: ows,
	}
}

func (ov *OwnerValidator) Validate(c *Config) error {
	for _, device := range c.Devices {
		if len(device.User) > 0 {
			_, err := ov.ownerService.GetUser(device.User)
			if err != nil {
				return err
			}
		}
		if len(device.Group) > 0 {
			_, err := ov.ownerService.GetGroup(device.Group)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type ResizeThresholdValidator struct{}

func NewResizeThresholdValidator() *ResizeThresholdValidator {
	return &ResizeThresholdValidator{}
}

func (rtv *ResizeThresholdValidator) Validate(c *Config) error {
	if !rtv.isValid(c.Defaults.ResizeThreshold) {
		return fmt.Errorf("🔴 '%g' (default) must be a floating point between 0 and 100 (inclusive)", c.Defaults.ResizeThreshold)
	}
	if !rtv.isValid(c.overrides.ResizeThreshold) {
		return fmt.Errorf("🔴 '%g' (-resize-threshold) must be a floating point between 0 and 100 (inclusive)", c.overrides.ResizeThreshold)
	}
	for name, device := range c.Devices {
		if !rtv.isValid(device.ResizeThreshold) {
			return fmt.Errorf("🔴 %s: '%g' must be a floating point between 0 and 100 (inclusive)", name, device.ResizeThreshold)
		}
	}
	return nil
}

func (rtv *ResizeThresholdValidator) isValid(rt float64) bool {
	return rt >= 0 && rt <= 100
}

type LvmConsumptionValidator struct{}

func NewLvmConsumptionValidator() *LvmConsumptionValidator {
	return &LvmConsumptionValidator{}
}

func (lcv *LvmConsumptionValidator) Validate(c *Config) error {
	if !lcv.isValid(c.Defaults.LvmConsumption) {
		return fmt.Errorf("🔴 '%d' (default) must be an integer between 0 and 100 (inclusive)", c.Defaults.LvmConsumption)
	}
	if !lcv.isValid(c.overrides.LvmConsumption) {
		return fmt.Errorf("🔴 '%d' (-lvm-consumption) must be an integer between 0 and 100 (inclusive)", c.overrides.LvmConsumption)
	}
	for name, device := range c.Devices {
		if !lcv.isValid(device.LvmConsumption) {
			return fmt.Errorf("🔴 %s: '%d' must be an integer between 0 and 100 (inclusive)", name, device.LvmConsumption)
		}
	}
	return nil
}

func (lcv *LvmConsumptionValidator) isValid(lc int) bool {
	return lc >= 0 && lc <= 100
}
