package state

import (
	"fmt"
	"log"
	"ebs-bootstrap/internal/config"
	"ebs-bootstrap/internal/service"
)

type deviceProperties struct {
	Name			string
	Fs     			string
	MountPoint  	string
	Owner  			string
	Group  			string
	Label  			string
	Permissions 	string
}

type Device struct {
	Properties		deviceProperties
	DeviceService	service.DeviceService
	FileService 	service.FileService
}

func NewDevice(name string, ds service.DeviceService,  fs service.FileService) (*Device, error) {
	s := &Device{
		DeviceService: ds,
		FileService: fs,
		Properties: deviceProperties{Name: name},
	}
	err := s.Pull()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (d *Device) Pull() (error) {
	name := d.Properties.Name
	di, err := d.DeviceService.GetDeviceInfo(name)
	if err != nil {
		return err
	}
	p := deviceProperties{
		Name: name,
		Fs: di.Fs,
		Label: di.Label,
		MountPoint: di.MountPoint,
	}

	if p.MountPoint == "" {
		log.Printf("🟡 %s: No mount-point detected. Skip further checks...", name)
		d.Properties = p
		return nil
	}

	fi, err := d.FileService.GetStats(p.MountPoint)
	if err != nil {
		return err
	}
	if fi.Exists {
		p.Owner = fi.Owner
		p.Group = fi.Group
		p.Permissions = fi.Permissions
	}

	d.Properties = p
	return nil
}

func (d *Device) Diff(c *config.Config) (error) {
	name := d.Properties.Name
	if name == "" {
		return fmt.Errorf("🔴 An unexpected error occured")
	}
	desired, found := c.Devices[name]
	if !found {
		return fmt.Errorf("🔴 %s: Couldn't find device in config", name)
	}

	if d.Properties.Fs != string(desired.Fs) {
		return fmt.Errorf("🔴 File System [%s]: Expected=%s", d.Properties.Name, desired.Fs)
	}

	if d.Properties.Label != string(desired.Label) {
		return fmt.Errorf("🔴 Label [%s]: Expected=%s", d.Properties.Name, desired.Label)
	}

	if d.Properties.MountPoint != string(desired.MountPoint) {
		return fmt.Errorf("🔴 Mount Point [%s]: Expected=%s", d.Properties.Name, desired.MountPoint)
	}

	if d.Properties.Owner != string(desired.Owner) {
		return fmt.Errorf("🔴 Owner [%s]: Expected=%s", d.Properties.MountPoint, desired.Owner)
	}

	if d.Properties.Group != string(desired.Group) {
		return fmt.Errorf("🔴 Group: [%s]: Expected=%s", d.Properties.MountPoint, desired.Group)
	}

	if d.Properties.Permissions != string(desired.Permissions) {
		return fmt.Errorf("🔴 Permissions [%s]: Expected=%s", d.Properties.MountPoint, desired.Permissions)
	}
	return nil
}

func (d *Device) Push(c *config.Config) (error) {
	return nil
}
