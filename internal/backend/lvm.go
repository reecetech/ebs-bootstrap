package backend

import (
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
	GetVolumeGroups(name string) []*model.VolumeGroup
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
		db.lvmGraph.AddLogicalVolume(lv.Name, lv.VolumeGroup)
	}
	return nil
}
