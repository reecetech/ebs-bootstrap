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
	CreateLogicalVolume(name string, volumeGroup string, freeSpacePercent int) action.Action
	ActivateLogicalVolume(name string, volumeGroup string) action.Action
	GetVolumeGroups(name string) []*model.VolumeGroup
	GetLogicalVolume(name string, volumeGroup string) (*model.LogicalVolume, error)
	SearchLogicalVolumes(volumeGroup string) []*model.LogicalVolume
	SearchVolumeGroup(physicalVolume string) *model.VolumeGroup
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

func (lb *LinuxLvmBackend) GetVolumeGroups(name string) []*model.VolumeGroup {
	vgs := []*model.VolumeGroup{}
	node, err := lb.lvmGraph.GetVolumeGroup(name)
	if err != nil {
		return vgs
	}
	pvn := lb.lvmGraph.GetParents(node, datastructures.PhysicalVolume)
	for _, pv := range pvn {
		vgs = append(vgs, &model.VolumeGroup{
			Name:           node.Name,
			PhysicalVolume: pv.Name,
		})
	}
	return vgs
}

func (lb *LinuxLvmBackend) GetLogicalVolume(name string, volumeGroup string) (*model.LogicalVolume, error) {
	node, err := lb.lvmGraph.GetLogicalVolume(name, volumeGroup)
	if err != nil {
		return nil, err
	}
	vgs := lb.lvmGraph.GetParents(node, datastructures.VolumeGroup)
	if len(vgs) == 0 {
		return nil, fmt.Errorf("🔴 %s: Logical volume has no volume group", node.Name)
	}
	return &model.LogicalVolume{
		Name:        node.Name,
		VolumeGroup: vgs[0].Name,
		State:       model.LogicalVolumeState(node.State),
	}, nil
}

func (lb *LinuxLvmBackend) SearchLogicalVolumes(volumeGroup string) []*model.LogicalVolume {
	lvs := []*model.LogicalVolume{}
	node, err := lb.lvmGraph.GetVolumeGroup(volumeGroup)
	if err != nil {
		return lvs
	}
	lvn := lb.lvmGraph.GetChildren(node, datastructures.LogicalVolume)
	for _, lv := range lvn {
		lvs = append(lvs, &model.LogicalVolume{
			Name:        lv.Name,
			VolumeGroup: node.Name,
			State:       model.LogicalVolumeState(lv.State),
		})
	}
	return lvs
}

func (lb *LinuxLvmBackend) SearchVolumeGroup(physicalVolume string) *model.VolumeGroup {
	node, err := lb.lvmGraph.GetPhysicalVolume(physicalVolume)
	if err != nil {
		return nil
	}
	vgn := lb.lvmGraph.GetChildren(node, datastructures.VolumeGroup)
	if len(vgn) == 0 {
		return nil
	}
	return &model.VolumeGroup{
		Name:           vgn[0].Name,
		PhysicalVolume: node.Name,
	}
}

func (lb *LinuxLvmBackend) CreatePhysicalVolume(name string) action.Action {
	return action.NewCreatePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateVolumeGroup(name string, physicalVolume string) action.Action {
	return action.NewCreateVolumeGroupAction(name, physicalVolume, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateLogicalVolume(name string, volumeGroup string, freeSpacePercent int) action.Action {
	return action.NewCreateLogicalVolumeAction(name, freeSpacePercent, volumeGroup, lb.lvmService)
}

func (lb *LinuxLvmBackend) ActivateLogicalVolume(name string, volumeGroup string) action.Action {
	return action.NewActivateLogicalVolumeAction(name, volumeGroup, lb.lvmService)
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
	lvs, err := db.lvmService.GetLogicalVolumes()
	if err != nil {
		return err
	}
	for _, lv := range lvs {
		db.lvmGraph.AddLogicalVolume(lv.Name, lv.VolumeGroup, datastructures.LvmNodeState(lv.State))
	}
	return nil
}
