package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type LvmService interface {
	GetDevices() ([]*model.Device, error)
	GetPhysicalVolumes() ([]*model.PhysicalVolume, error)
	GetVolumeGroups() ([]*model.VolumeGroup, error)
	GetLogicalVolumes() ([]*model.LogicalVolume, error)
	CreatePhysicalVolume(name string) error
	CreateVolumeGroup(name string, physicalVolume string) error
	CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) error
	ActivateLogicalVolume(name string, volumeGroup string) error
	ResizePhysicalVolume(name string) error
	ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) error
}

type LinuxLvmService struct {
	runnerFactory utils.RunnerFactory
}

type PvsResponse struct {
	Report []struct {
		PhysicalVolume []struct {
			Name               string `json:"pv_name"`
			PhysicalVolumeSize string `json:"pv_size"`
			DeviceSize         string `json:"dev_size"`
		} `json:"pv"`
	} `json:"report"`
}

type VgsResponse struct {
	Report []struct {
		VolumeGroup []struct {
			Name           string `json:"vg_name"`
			PhysicalVolume string `json:"pv_name"`
			Size           string `json:"vg_size"`
		} `json:"vg"`
	} `json:"report"`
}

type LvsResponse struct {
	Report []struct {
		LogicalVolume []struct {
			Name        string `json:"lv_name"`
			VolumeGroup string `json:"vg_name"`
			Attributes  string `json:"lv_attr"`
			Size        string `json:"lv_size"`
		} `json:"lv"`
	} `json:"report"`
}

func NewLinuxLvmService(rf utils.RunnerFactory) *LinuxLvmService {
	return &LinuxLvmService{
		runnerFactory: rf,
	}
}

func (ls *LinuxLvmService) GetDevices() ([]*model.Device, error) {
	r := ls.runnerFactory.Select(utils.Pvs)
	output, err := r.Command("-o", "pv_name,dev_size", "--reportformat", "json", "--units", "b", "--nosuffix")
	if err != nil {
		return nil, err
	}
	pr := &PvsResponse{}
	err = json.Unmarshal([]byte(output), pr)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode pvs response: %v", err)
	}
	pvs := make([]*model.Device, len(pr.Report[0].PhysicalVolume))
	for i, pv := range pr.Report[0].PhysicalVolume {
		size, err := strconv.ParseUint(pv.DeviceSize, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("🔴 Failed to cast device size to unsigned 64-bit integer")
		}
		pvs[i] = &model.Device{
			Name: pv.Name,
			Size: size,
		}
	}
	return pvs, nil
}

func (ls *LinuxLvmService) GetPhysicalVolumes() ([]*model.PhysicalVolume, error) {
	r := ls.runnerFactory.Select(utils.Pvs)
	output, err := r.Command("-o", "pv_name,pv_size", "--reportformat", "json", "--units", "b", "--nosuffix")
	if err != nil {
		return nil, err
	}
	pr := &PvsResponse{}
	err = json.Unmarshal([]byte(output), pr)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode pvs response: %v", err)
	}
	pvs := make([]*model.PhysicalVolume, len(pr.Report[0].PhysicalVolume))
	for i, pv := range pr.Report[0].PhysicalVolume {
		size, err := strconv.ParseUint(pv.PhysicalVolumeSize, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("🔴 Failed to cast physical volume size to unsigned 64-bit integer")
		}
		pvs[i] = &model.PhysicalVolume{
			Name: pv.Name,
			Size: size,
		}
	}
	return pvs, nil
}

func (ls *LinuxLvmService) GetVolumeGroups() ([]*model.VolumeGroup, error) {
	r := ls.runnerFactory.Select(utils.Vgs)
	output, err := r.Command("-o", "vg_name,pv_name,vg_size", "--reportformat", "json", "--units", "b", "--nosuffix")
	if err != nil {
		return nil, err
	}
	vr := &VgsResponse{}
	err = json.Unmarshal([]byte(output), vr)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode vgs response: %v", err)
	}
	vgs := make([]*model.VolumeGroup, len(vr.Report[0].VolumeGroup))
	for i, vg := range vr.Report[0].VolumeGroup {
		size, err := strconv.ParseUint(vg.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("🔴 Failed to cast volume group size to unsigned 64-bit integer")
		}
		vgs[i] = &model.VolumeGroup{
			Name:           vg.Name,
			PhysicalVolume: vg.PhysicalVolume,
			State:          model.VolumeGroupInactive,
			Size:           size,
		}
	}
	return vgs, nil
}

func (ls *LinuxLvmService) GetLogicalVolumes() ([]*model.LogicalVolume, error) {
	r := ls.runnerFactory.Select(utils.Lvs)
	output, err := r.Command("-o", "lv_name,vg_name,lv_attr,lv_size", "--reportformat", "json", "--units", "b", "--nosuffix")
	if err != nil {
		return nil, err
	}
	lr := &LvsResponse{}
	err = json.Unmarshal([]byte(output), lr)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode lvs response: %v", err)
	}
	lvs := make([]*model.LogicalVolume, len(lr.Report[0].LogicalVolume))
	for i, lv := range lr.Report[0].LogicalVolume {
		// Get Logical Volume State
		var state model.LvmState
		switch lv.Attributes[4] {
		case 'a':
			state = model.LogicalVolumeActive
		case '-':
			state = model.LogicalVolumeInactive
		default:
			state = model.LogicalVolumeUnsupported
		}

		// Get Logical Volume Size
		size, err := strconv.ParseUint(lv.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("🔴 Failed to cast logical volume size to unsigned 64-bit integer")
		}

		lvs[i] = &model.LogicalVolume{
			Name:        lv.Name,
			VolumeGroup: lv.VolumeGroup,
			State:       state,
			Size:        size,
		}
	}
	return lvs, nil
}

func (ls *LinuxLvmService) CreatePhysicalVolume(name string) error {
	r := ls.runnerFactory.Select(utils.PvCreate)
	_, err := r.Command(name)
	return err
}

func (ls *LinuxLvmService) CreateVolumeGroup(name string, physicalVolume string) error {
	r := ls.runnerFactory.Select(utils.VgCreate)
	_, err := r.Command(name, physicalVolume)
	return err
}

func (ls *LinuxLvmService) CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) error {
	r := ls.runnerFactory.Select(utils.LvCreate)
	_, err := r.Command("-l", fmt.Sprintf("%d%%VG", volumeGroupPercent), "-n", name, volumeGroup)
	return err
}

func (ls *LinuxLvmService) ActivateLogicalVolume(name string, volumeGroup string) error {
	r := ls.runnerFactory.Select(utils.LvChange)
	_, err := r.Command("-ay", fmt.Sprintf("%s/%s", volumeGroup, name))
	return err
}

func (ls *LinuxLvmService) ResizePhysicalVolume(name string) error {
	r := ls.runnerFactory.Select(utils.PvResize)
	_, err := r.Command(name)
	return err
}

func (ls *LinuxLvmService) ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) error {
	r := ls.runnerFactory.Select(utils.LvExtend)
	_, err := r.Command("-l", fmt.Sprintf("%d%%VG", volumeGroupPercent), fmt.Sprintf("%s/%s", volumeGroup, name))
	return err
}
