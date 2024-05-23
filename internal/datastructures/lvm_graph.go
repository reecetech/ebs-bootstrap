package datastructures

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type LvmNode struct {
	id       string
	Name     string
	State    model.LvmState
	Size     uint64
	children []*LvmNode
	parents  []*LvmNode
}

func NewDevice(name string, size uint64) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("device:%s", name),
		Name:     name,
		State:    model.DeviceActive,
		Size:     size,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewPhysicalVolume(name string, size uint64) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("pv:%s", name),
		Name:     name,
		State:    model.PhysicalVolumeActive,
		Size:     size,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewVolumeGroup(name string, size uint64) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("vg:%s", name),
		Name:     name,
		State:    model.VolumeGroupInactive,
		Size:     size,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewLogicalVolume(name string, vg string, State model.LvmState, size uint64) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("lv:%s:vg:%s", name, vg),
		Name:     name,
		State:    State,
		Size:     size,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

type LvmGraph struct {
	nodes map[string]*LvmNode
}

func NewLvmGraph() *LvmGraph {
	return &LvmGraph{
		nodes: map[string]*LvmNode{},
	}
}

func (lg *LvmGraph) AddDevice(name string, size uint64) error {
	bd := NewDevice(name, size)

	_, found := lg.nodes[bd.id]
	if found {
		return fmt.Errorf("🔴 %s: Device already exists", name)
	}

	lg.nodes[bd.id] = bd
	return nil
}

func (lg *LvmGraph) AddPhysicalVolume(name string, size uint64) error {
	pv := NewPhysicalVolume(name, size)

	_, found := lg.nodes[pv.id]
	if found {
		return fmt.Errorf("🔴 %s: Physical volume already exists", name)
	}

	dn, err := lg.GetDevice(name)
	if err != nil {
		return err
	}

	lg.nodes[pv.id] = pv
	dn.children = append(dn.children, pv)
	pv.parents = append(pv.parents, dn)
	return nil
}

func (lg *LvmGraph) AddVolumeGroup(name string, pv string, size uint64) error {
	id := fmt.Sprintf("vg:%s", name)

	vgn, found := lg.nodes[id]
	if !found {
		vgn = NewVolumeGroup(name, size)
	}

	pvn, err := lg.GetPhysicalVolume(pv)
	if err != nil {
		return err
	}

	if len(pvn.children) > 0 {
		return fmt.Errorf("🔴 %s: Physical volume is already assigned to volume group %s", pv, pvn.children[0].Name)
	}

	lg.nodes[vgn.id] = vgn
	pvn.children = append(pvn.children, vgn)
	vgn.parents = append(vgn.parents, pvn)
	return nil
}

func (lg *LvmGraph) AddLogicalVolume(name string, vg string, state model.LvmState, size uint64) error {
	lvn := NewLogicalVolume(name, vg, state, size)

	_, found := lg.nodes[lvn.id]
	if found {
		return fmt.Errorf("🔴 %s/%s: Logical volume already exists", name, vg)
	}

	vgn, err := lg.GetVolumeGroup(vg)
	if err != nil {
		return err
	}

	lg.nodes[lvn.id] = lvn
	vgn.children = append(vgn.children, lvn)
	lvn.parents = append(lvn.parents, vgn)

	// If at least one of the logical volumes are active, the
	// volume group is considered active
	for _, lvn := range vgn.children {
		if lvn.State == model.LogicalVolumeActive {
			vgn.State = model.VolumeGroupActive
			break
		}
	}
	return nil
}

func (lg *LvmGraph) GetDevice(name string) (*LvmNode, error) {
	id := fmt.Sprintf("device:%s", name)
	dn, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("🔴 %s: Block device does not exist", name)
	}
	return dn, nil
}

func (lg *LvmGraph) GetPhysicalVolume(name string) (*LvmNode, error) {
	id := fmt.Sprintf("pv:%s", name)
	pvn, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("🔴 %s: Physical volume does not exist", name)
	}
	return pvn, nil
}

func (lg *LvmGraph) GetVolumeGroup(name string) (*LvmNode, error) {
	id := fmt.Sprintf("vg:%s", name)
	vgn, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("🔴 %s: Volume group does not exist", name)
	}
	return vgn, nil
}

func (lg *LvmGraph) GetLogicalVolume(name string, vg string) (*LvmNode, error) {
	id := fmt.Sprintf("lv:%s:vg:%s", name, vg)
	lvn, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("🔴 %s/%s: Logical volume does not exist", vg, name)
	}
	return lvn, nil
}

func (lg *LvmGraph) GetParents(node *LvmNode, kind model.LvmKind) []*LvmNode {
	parents := []*LvmNode{}
	for _, p := range node.parents {
		// Bitmasking to check if the parent nodes is of the desired kind
		if int32(p.State)&int32(kind) > 0 {
			parents = append(parents, p)
		}
	}
	return parents
}

func (lg *LvmGraph) GetChildren(node *LvmNode, kind model.LvmKind) []*LvmNode {
	children := []*LvmNode{}
	for _, c := range node.children {
		// Bitmasking to check if children nodes is of the desired kind
		if int32(c.State)&int32(kind) > 0 {
			children = append(children, c)
		}
	}
	return children
}
