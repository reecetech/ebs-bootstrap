package service

import (
	"encoding/json"
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type LvmService interface {
	GetPhysicalVolumes() ([]*model.PhysicalVolume, error)
	GetVolumeGroups() ([]*model.VolumeGroup, error)
	CreatePhysicalVolume(name string) error
	CreateVolumeGroup(name string, physicalVolume string) error
}

type LinuxLvmService struct {
	runnerFactory utils.RunnerFactory
}

type PvsResponse struct {
	Report []struct {
		PhysicalVolume []struct {
			Name string `json:"pv_name"`
		} `json:"pv"`
	} `json:"report"`
}

type VgsResponse struct {
	Report []struct {
		VolumeGroup []struct {
			Name           string `json:"vg_name"`
			PhysicalVolume string `json:"pv_name"`
		} `json:"vg"`
	} `json:"report"`
}

func NewLinuxLvmService(rf utils.RunnerFactory) *LinuxLvmService {
	return &LinuxLvmService{
		runnerFactory: rf,
	}
}

func (ls *LinuxLvmService) GetPhysicalVolumes() ([]*model.PhysicalVolume, error) {
	r := ls.runnerFactory.Select(utils.Pvs)
	output, err := r.Command("-o", "pv_name", "--reportformat", "json")
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
		pvs[i] = &model.PhysicalVolume{
			Name: pv.Name,
		}
	}
	return pvs, nil
}

func (ls *LinuxLvmService) GetVolumeGroups() ([]*model.VolumeGroup, error) {
	r := ls.runnerFactory.Select(utils.Vgs)
	output, err := r.Command("-o", "vg_name,pv_name", "--reportformat", "json")
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
		vgs[i] = &model.VolumeGroup{
			Name:           vg.Name,
			PhysicalVolume: vg.PhysicalVolume,
		}
	}
	return vgs, nil
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
