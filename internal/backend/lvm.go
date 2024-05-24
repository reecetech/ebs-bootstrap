package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/datastructures"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

const (
	// The % tolerance to expect the logical volume size to be within
	// -------------------------------------------------------
	// If the (logical volume size / volume group size) * 100 is less than
	// (lvmConsumption% - tolerance%) then we perform a resize operation
	// -------------------------------------------------------
	// If the (logical volume / volume group size) * 100 is greater than
	// (lvmConsumption% + tolerance%) then the user is attempting a downsize
	// operation. We outright deny this as downsizing can be a destructive
	// operation
	// -------------------------------------------------------
	// Why implement a tolernace-based policy for resizing?
	// 	- When creating a Logical Volume, `ebs-bootstrap` issues a command like
	// 		`lvcreate -l 20%VG -n lv_name vg_name`
	// 	- When we calculate how much percentage of the volume group has been
	// 		consumed by the logical volume, the value would look like 20.0052096...
	// 	- A tolerance establishes a window of acceptable values for avoiding a
	// 		resizing operation
	LogicalVolumeResizeTolerance = float64(0.1)
	// The % threshold at which to resize a physical volume
	// -------------------------------------------------------
	// If the (physical volume size / device size) * 100 falls
	// under this threshold then we perform a resize operation
	// -------------------------------------------------------
	// The smallest gp3 EBS volume you can create is 1GiB (1073741824 bytes).
	// The default size of the extent of a PV is 4 MiB (4194304 bytes).
	// Typically, the first extent of a PV is reserved for metadata. This
	// produces a PV of size 1069547520 bytes (Usage=99.6093%). We ensure
	// that we set the resize threshold to 99.6% to ensure that a 1 GiB EBS
	// volume won't be always resized
	// -------------------------------------------------------
	// Why not just look for a difference of 4194304 bytes?
	//	- The size of the extent can be changed by the user
	//	- Therefore we may not always see a difference of 4194304 bytes between
	//	  the block device and physical volume size
	PhysicalVolumeResizeThreshold = float64(99.6)
)

type LvmBackend interface {
	CreatePhysicalVolume(name string) action.Action
	CreateVolumeGroup(name string, physicalVolume string) action.Action
	CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) action.Action
	ActivateLogicalVolume(name string, volumeGroup string) action.Action
	GetVolumeGroups(name string) []*model.VolumeGroup
	GetLogicalVolume(name string, volumeGroup string) (*model.LogicalVolume, error)
	SearchLogicalVolumes(volumeGroup string) ([]*model.LogicalVolume, error)
	SearchVolumeGroup(physicalVolume string) (*model.VolumeGroup, error)
	ShouldResizePhysicalVolume(name string) (bool, error)
	ResizePhysicalVolume(name string) action.Action
	ShouldResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) (bool, error)
	ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) action.Action
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
	vgn, err := lb.lvmGraph.GetVolumeGroup(name)
	if err != nil {
		return vgs
	}
	pvn := lb.lvmGraph.GetParents(vgn, model.PhysicalVolumeKind)
	for _, pv := range pvn {
		vgs = append(vgs, &model.VolumeGroup{
			Name:           vgn.Name,
			PhysicalVolume: pv.Name,
			State:          vgn.State,
			Size:           vgn.Size,
		})
	}
	return vgs
}

func (lb *LinuxLvmBackend) GetLogicalVolume(name string, volumeGroup string) (*model.LogicalVolume, error) {
	lvn, err := lb.lvmGraph.GetLogicalVolume(name, volumeGroup)
	if err != nil {
		return nil, err
	}
	vgs := lb.lvmGraph.GetParents(lvn, model.VolumeGroupKind)
	if len(vgs) == 0 {
		return nil, fmt.Errorf("🔴 %s: Logical volume has no volume group", lvn.Name)
	}
	return &model.LogicalVolume{
		Name:        lvn.Name,
		VolumeGroup: vgs[0].Name,
		State:       lvn.State,
		Size:        lvn.Size,
	}, nil
}

func (lb *LinuxLvmBackend) SearchLogicalVolumes(volumeGroup string) ([]*model.LogicalVolume, error) {
	lvs := []*model.LogicalVolume{}
	vgn, err := lb.lvmGraph.GetVolumeGroup(volumeGroup)
	if err != nil {
		return nil, err
	}
	lvns := lb.lvmGraph.GetChildren(vgn, model.LogicalVolumeKind)
	for _, lvn := range lvns {
		lvs = append(lvs, &model.LogicalVolume{
			Name:        lvn.Name,
			VolumeGroup: vgn.Name,
			State:       lvn.State,
			Size:        lvn.Size,
		})
	}
	return lvs, nil
}

func (lb *LinuxLvmBackend) SearchVolumeGroup(physicalVolume string) (*model.VolumeGroup, error) {
	pvn, err := lb.lvmGraph.GetPhysicalVolume(physicalVolume)
	if err != nil {
		return nil, err
	}
	vgn := lb.lvmGraph.GetChildren(pvn, model.VolumeGroupKind)
	if len(vgn) == 0 {
		return nil, fmt.Errorf("🔴 %s: Physical volume has no volume group", physicalVolume)
	}
	return &model.VolumeGroup{
		Name:           vgn[0].Name,
		PhysicalVolume: pvn.Name,
		State:          vgn[0].State,
		Size:           vgn[0].Size,
	}, nil
}

func (lb *LinuxLvmBackend) CreatePhysicalVolume(name string) action.Action {
	return action.NewCreatePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateVolumeGroup(name string, physicalVolume string) action.Action {
	return action.NewCreateVolumeGroupAction(name, physicalVolume, lb.lvmService)
}

func (lb *LinuxLvmBackend) CreateLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) action.Action {
	return action.NewCreateLogicalVolumeAction(name, volumeGroupPercent, volumeGroup, lb.lvmService)
}

func (lb *LinuxLvmBackend) ActivateLogicalVolume(name string, volumeGroup string) action.Action {
	return action.NewActivateLogicalVolumeAction(name, volumeGroup, lb.lvmService)
}

func (lb *LinuxLvmBackend) ShouldResizePhysicalVolume(name string) (bool, error) {
	pvn, err := lb.lvmGraph.GetPhysicalVolume(name)
	if err != nil {
		return false, nil
	}
	dn := lb.lvmGraph.GetParents(pvn, model.DeviceKind)
	if len(dn) == 0 {
		return false, nil
	}
	return (float64(pvn.Size) / float64(dn[0].Size) * 100) < PhysicalVolumeResizeThreshold, nil
}

func (lb *LinuxLvmBackend) ResizePhysicalVolume(name string) action.Action {
	return action.NewResizePhysicalVolumeAction(name, lb.lvmService)
}

func (lb *LinuxLvmBackend) ShouldResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) (bool, error) {
	left := float64(volumeGroupPercent) - LogicalVolumeResizeTolerance
	right := float64(volumeGroupPercent) + LogicalVolumeResizeTolerance
	lvn, err := lb.lvmGraph.GetLogicalVolume(name, volumeGroup)
	if err != nil {
		return false, err
	}
	vgn := lb.lvmGraph.GetParents(lvn, model.VolumeGroupKind)
	if len(vgn) == 0 {
		return false, fmt.Errorf("🔴 %s: Logical volume has no volume group", name)
	}
	usedPerecent := (float64(lvn.Size) / float64(vgn[0].Size)) * 100
	if usedPerecent > right {
		return false, fmt.Errorf("🔴 %s: Logical volume %s is using %.0f%% of volume group %s, which exceeds the expected usage of %d%%", volumeGroup, name, usedPerecent, volumeGroup, volumeGroupPercent)
	}
	return usedPerecent < left, nil
}

func (lb *LinuxLvmBackend) ResizeLogicalVolume(name string, volumeGroup string, volumeGroupPercent uint64) action.Action {
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
		err := lvmGraph.AddLogicalVolume(lv.Name, lv.VolumeGroup, lv.State, lv.Size)
		if err != nil {
			return err
		}
	}

	db.lvmGraph = lvmGraph
	return nil
}
