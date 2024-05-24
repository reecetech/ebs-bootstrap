package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/datastructures"
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
	ShouldResizePhysicalVolume(name string, threshold float64) bool
	ResizePhysicalVolume(name string) action.Action
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
			Size:        lv.Size,
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
		Size:           vgn[0].Size,
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

func (lb *LinuxLvmBackend) ShouldResizePhysicalVolume(name string, threshold float64) bool {
	node, err := lb.lvmGraph.GetPhysicalVolume(name)
	if err != nil {
		return false
	}
	dvn := lb.lvmGraph.GetParents(node, datastructures.Device)
	if len(dvn) == 0 {
		return false
	}
	return (float64(node.Size) / float64(dvn[0].Size) * 100) < threshold
}

func (lb *LinuxLvmBackend) ResizePhysicalVolume(name string) action.Action {
	return action.NewResizePhysicalVolumeAction(name, lb.lvmService)
}

func (db *LinuxLvmBackend) From(config *config.Config) error {
	ds, err := db.lvmService.GetDevices()
	if err != nil {
		return err
	}
	for _, d := range ds {
		db.lvmGraph.AddDevice(d.Name, d.Size)
	}
	pvs, err := db.lvmService.GetPhysicalVolumes()
	if err != nil {
		return err
	}
	for _, pv := range pvs {
		db.lvmGraph.AddPhysicalVolume(pv.Name, pv.Size)
	}
	vgs, err := db.lvmService.GetVolumeGroups()
	if err != nil {
		return err
	}
	for _, vg := range vgs {
		db.lvmGraph.AddVolumeGroup(vg.Name, vg.PhysicalVolume, vg.Size)
	}
	lvs, err := db.lvmService.GetLogicalVolumes()
	if err != nil {
		return err
	}
	for _, lv := range lvs {
		db.lvmGraph.AddLogicalVolume(lv.Name, lv.VolumeGroup, datastructures.LvmNodeState(lv.State), lv.Size)
	}
	return nil
}
