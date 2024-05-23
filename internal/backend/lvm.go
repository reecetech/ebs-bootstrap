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
	CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) action.Action
	ActivateLogicalVolume(name string, volumeGroup string) action.Action
	GetVolumeGroups(name string) []*model.VolumeGroup
	GetLogicalVolume(name string, volumeGroup string) (*model.LogicalVolume, error)
	SearchLogicalVolumes(volumeGroup string) ([]*model.LogicalVolume, error)
	SearchVolumeGroup(physicalVolume string) (*model.VolumeGroup, error)
	ShouldResizePhysicalVolume(name string, threshold float64) (bool, error)
	ResizePhysicalVolume(name string) action.Action
	ShouldResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int, tolerance float64) (bool, error)
	ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) action.Action
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

func (lb *LinuxLvmBackend) SearchLogicalVolumes(volumeGroup string) ([]*model.LogicalVolume, error) {
	lvs := []*model.LogicalVolume{}
	node, err := lb.lvmGraph.GetVolumeGroup(volumeGroup)
	if err != nil {
		return nil, err
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
	return lvs, nil
}

func (lb *LinuxLvmBackend) SearchVolumeGroup(physicalVolume string) (*model.VolumeGroup, error) {
	node, err := lb.lvmGraph.GetPhysicalVolume(physicalVolume)
	if err != nil {
		return nil, err
	}
	vgn := lb.lvmGraph.GetChildren(node, datastructures.VolumeGroup)
	if len(vgn) == 0 {
		return nil, fmt.Errorf("🔴 %s: Physical volume has no volume group", physicalVolume)
	}
	return &model.VolumeGroup{
		Name:           vgn[0].Name,
		PhysicalVolume: node.Name,
		Size:           vgn[0].Size,
	}, nil
}

func (lb *LinuxLvmBackend) CreatePhysicalVolume(name string) action.Action {
	return action.NewCreatePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateVolumeGroup(name string, physicalVolume string) action.Action {
	return action.NewCreateVolumeGroupAction(name, physicalVolume, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) action.Action {
	return action.NewCreateLogicalVolumeAction(name, volumeGroupPercent, volumeGroup, lb.lvmService)
}

func (lb *LinuxLvmBackend) ActivateLogicalVolume(name string, volumeGroup string) action.Action {
	return action.NewActivateLogicalVolumeAction(name, volumeGroup, lb.lvmService)
}

func (lb *LinuxLvmBackend) ShouldResizePhysicalVolume(name string, threshold float64) (bool, error) {
	node, err := lb.lvmGraph.GetPhysicalVolume(name)
	if err != nil {
		return false, nil
	}
	dvn := lb.lvmGraph.GetParents(node, datastructures.Device)
	if len(dvn) == 0 {
		return false, nil
	}
	return (float64(node.Size) / float64(dvn[0].Size) * 100) < threshold, nil
}

func (lb *LinuxLvmBackend) ResizePhysicalVolume(name string) action.Action {
	return action.NewResizePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) ShouldResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int, tolerance float64) (bool, error) {
	left := float64(volumeGroupPercent) - tolerance
	right := float64(volumeGroupPercent) + tolerance
	node, err := lb.lvmGraph.GetLogicalVolume(name, volumeGroup)
	if err != nil {
		return false, err
	}
	vgn := lb.lvmGraph.GetParents(node, datastructures.VolumeGroup)
	if len(vgn) == 0 {
		return false, fmt.Errorf("🔴 %s: Logical volume has no volume group", name)
	}
	usedPerecent := (float64(node.Size) / float64(vgn[0].Size)) * 100
	if usedPerecent > right {
		return false, fmt.Errorf("🔴 %s: Logical volume %s is using %.0f%% of volume group %s, which exceeds the expected usage of %d%%", volumeGroup, name, usedPerecent, volumeGroup, volumeGroupPercent)
	}
	return usedPerecent < left, nil
}

func (lb *LinuxLvmBackend) ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent int) action.Action {
	return action.NewResizeLogicalVolumeAction(name, volumeGroupPercent, volumeGroup, lb.lvmService)
}

func (db *LinuxLvmBackend) From(config *config.Config) error {
	// We populate a temporary lvmGraph and then assign it to the backend
	// after all objects have been successfully added. This avoids a partial
	// state in the event of failure during one of intermediate steps.
	db.lvmGraph = nil
	lvmGraph := datastructures.NewLvmGraph()

	ds, err := db.lvmService.GetDevices()
	if err != nil {
		return err
	}
	for _, d := range ds {
		err := lvmGraph.AddDevice(d.Name, d.Size)
		if err != nil {
			return err
		}
	}
	pvs, err := db.lvmService.GetPhysicalVolumes()
	if err != nil {
		return err
	}
	for _, pv := range pvs {
		err := lvmGraph.AddPhysicalVolume(pv.Name, pv.Size)
		if err != nil {
			return err
		}
	}
	vgs, err := db.lvmService.GetVolumeGroups()
	if err != nil {
		return err
	}
	for _, vg := range vgs {
		err := lvmGraph.AddVolumeGroup(vg.Name, vg.PhysicalVolume, vg.Size)
		if err != nil {
			return err
		}
	}
	lvs, err := db.lvmService.GetLogicalVolumes()
	if err != nil {
		return err
	}
	for _, lv := range lvs {
		err := lvmGraph.AddLogicalVolume(lv.Name, lv.VolumeGroup, datastructures.LvmNodeState(lv.State), lv.Size)
		if err != nil {
			return err
		}
	}

	db.lvmGraph = lvmGraph
	return nil
}
