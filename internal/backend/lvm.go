package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	datastructures "github.com/reecetech/ebs-bootstrap/internal/data_structures"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type LvmBackend interface {
	CreatePhysicalVolume(name string) action.Action
	CreateVolumeGroup(name string, physicalVolume string) action.Action
	GetVolumeGroup(name string) (*model.VolumeGroup, error)
	SearchVolumeGroup(pv *model.PhysicalVolume) (*model.VolumeGroup, error)
	GetPhysicalVolume(name string) (*model.PhysicalVolume, error)
	SearchPhysicalVolumes(vg *model.VolumeGroup) ([]*model.PhysicalVolume, error)
	From(config *config.Config) error
}

type LinuxLvmBackend struct {
	lvmGraph   *datastructures.LvmGraph
	lvmService service.LvmService
}

func NewLinuxLvmBackend(ls service.LvmService) *LinuxLvmBackend {
	return &LinuxLvmBackend{
		lvmGraph:   datastructures.NewLvmGraph(),
		lvmService: ls,
	}
}

func (lb *LinuxLvmBackend) GetVolumeGroup(name string) (*model.VolumeGroup, error) {
	node, err := lb.lvmGraph.GetVolumeGroup(name)
	if err != nil {
		return nil, err
	}
	return &model.VolumeGroup{
		Name: node.Name,
	}, nil
}

func (lb *LinuxLvmBackend) SearchVolumeGroup(pv *model.PhysicalVolume) (*model.VolumeGroup, error) {
	pvn, err := lb.lvmGraph.GetPhysicalVolume(pv.Name)
	if err != nil {
		return nil, err
	}
	vg := lb.lvmGraph.GetChildren(pvn, datastructures.VolumeGroup)
	if len(vg) == 0 {
		return nil, nil
	}
	return &model.VolumeGroup{Name: vg[0].Name}, nil
}

func (lb *LinuxLvmBackend) GetPhysicalVolume(name string) (*model.PhysicalVolume, error) {
	node, err := lb.lvmGraph.GetPhysicalVolume(name)
	if err != nil {
		return nil, err
	}
	return &model.PhysicalVolume{
		Name: node.Name,
	}, nil
}

func (lb *LinuxLvmBackend) SearchPhysicalVolumes(vg *model.VolumeGroup) ([]*model.PhysicalVolume, error) {
	vgn, err := lb.lvmGraph.GetVolumeGroup(vg.Name)
	if err != nil {
		return nil, err
	}
	pvs := lb.lvmGraph.GetParents(vgn, datastructures.PhysicalVolume)
	physicalVolumes := make([]*model.PhysicalVolume, len(pvs))
	for i, pv := range pvs {
		physicalVolumes[i] = &model.PhysicalVolume{
			Name: pv.Name,
		}
	}
	if len(physicalVolumes) == 0 {
		return nil, fmt.Errorf("🔴 %s: No physical volumes found", vg.Name)
	}
	return physicalVolumes, nil
}

func (lb *LinuxLvmBackend) CreatePhysicalVolume(name string) action.Action {
	return action.NewCreatePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateVolumeGroup(name string, physicalVolume string) action.Action {
	return action.NewCreateVolumeGroupAction(name, physicalVolume, lb.lvmService)
}

func (db *LinuxLvmBackend) From(config *config.Config) error {
	pvs, err := db.lvmService.GetPhysicalVolumes()
	if err != nil {
		return err
	}
	for _, pv := range pvs {
		db.lvmGraph.AddBlockDevice(pv.Name)
		db.lvmGraph.AddPhysicalVolume(pv.Name)
	}
	vgs, err := db.lvmService.GetVolumeGroups()
	if err != nil {
		return err
	}
	for _, vg := range vgs {
		db.lvmGraph.AddVolumeGroup(vg.Name, vg.PhysicalVolume)
	}
	return nil
}
