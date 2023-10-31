package config

import (
	"fmt"
	"os/user"
	"strconv"
	"slices"
	"ebs-bootstrap/internal/service"
)

var (
	ValidDeviceModes = []string{"healthcheck"}
)

type Modifiers interface {
	Modify(config *Config) (error)
}

type OwnerModifier struct {}

func (om *OwnerModifier) Modify(config *Config) (error) {
	for key, device := range config.Devices {
		var u *user.User;
		var err error;
		if _, atoiErr := strconv.Atoi(device.Owner); atoiErr != nil {
			u, err = user.Lookup(device.Owner)
		} else {
			u, err = user.LookupId(device.Owner)
		}
		if err != nil {
			return err
		}
		device.Owner = u.Uid
		config.Devices[key] = device
	}
	return nil
}

type DeviceModifier struct {
	DeviceTranslator	*service.DeviceTranslator
}

func (dm *DeviceModifier) Modify(config *Config) (error) {
	for key, device := range config.Devices {
		alias, found := dm.DeviceTranslator.Table[key]
		if !found {
			return fmt.Errorf("🔴 Could not identify a device with an alias %s", key)
		}
		delete(config.Devices, key)
		config.Devices[alias] = device
	}
	return nil
}

type GroupModifier struct {}

func (gm *GroupModifier) Modify(config *Config) (error) {
	for key, device := range config.Devices {
		var g *user.Group;
		var err error;
		if _, atoiErr := strconv.Atoi(device.Group); atoiErr != nil {
			g, err = user.LookupGroup(device.Group)
		} else {
			g, err = user.LookupGroupId(device.Group)
		}
		if err != nil {
			return err
		}
		device.Group = g.Gid
		config.Devices[key] = device
	}
	return nil
}

type DeviceModeModifier struct {}

func (dm *DeviceModeModifier) Modify(config *Config) (error) {
	if config.Global.Mode != "" && !slices.Contains(ValidDeviceModes, config.Global.Mode) {
		return fmt.Errorf("🔴 A valid global mode was not provided: Expected=%s Provided=%s", ValidDeviceModes, config.Global.Mode)
	}

	for key, device := range config.Devices {
		if device.Mode == "" && config.Global.Mode == "" {
			return fmt.Errorf("🔴 %s: If mode is not provided locally, it must be provided globally", key)
		}

		if device.Mode != "" && !slices.Contains(ValidDeviceModes, device.Mode) {
			return fmt.Errorf("🔴 %s: A valid mode was not provided: Expected=%s Provided=%s", key, ValidDeviceModes, device.Mode)
		}

		if device.Mode == "" {
			device.Mode = config.Global.Mode
		}
		config.Devices[key] = device
	}
	return nil
}
