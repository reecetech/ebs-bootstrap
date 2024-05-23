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
	nodes map[string]*LvmNode // A map that stores all the nodes by their Id
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
		return fmt.Errorf("block device %s already exists", name)
	}

	lg.nodes[bd.id] = bd
	return nil
}

func (lg *LvmGraph) AddPhysicalVolume(name string, size uint64) error {
	pv := NewPhysicalVolume(name, size)

	_, found := lg.nodes[pv.id]
	if found {
		return fmt.Errorf("physical volume %s already exists", name)
	}

	bdId := fmt.Sprintf("device:%s", name)
	bdn, found := lg.nodes[bdId]
	if !found {
		return fmt.Errorf("block device %s does not exist", name)
	}

	lg.nodes[pv.id] = pv
	bdn.children = append(bdn.children, pv)
	pv.parents = append(pv.parents, bdn)
	return nil
}

func (lg *LvmGraph) AddVolumeGroup(name string, pv string, size uint64) error {
	id := fmt.Sprintf("vg:%s", name)

	vg, found := lg.nodes[id]
	if !found {
		vg = NewVolumeGroup(name, size)
	}

	pvId := fmt.Sprintf("pv:%s", pv)
	pvn, found := lg.nodes[pvId]
	if !found {
		return fmt.Errorf("physical volume %s does not exist", pv)
	}

	if len(pvn.children) > 0 {
		return fmt.Errorf("%s is already assigned to volume group %s", pv, pvn.children[0].Name)
	}

	lg.nodes[vg.id] = vg
	pvn.children = append(pvn.children, vg)
	vg.parents = append(vg.parents, pvn)
	return nil
}

func (lg *LvmGraph) AddLogicalVolume(name string, vg string, state model.LvmState, size uint64) error {
	lv := NewLogicalVolume(name, vg, state, size)

	_, found := lg.nodes[lv.id]
	if found {
		return fmt.Errorf("logical volume %s already exists", name)
	}

	vgId := fmt.Sprintf("vg:%s", vg)
	vgn, found := lg.nodes[vgId]
	if !found {
		return fmt.Errorf("volume group %s does not exist", vg)
	}

	lg.nodes[lv.id] = lv
	vgn.children = append(vgn.children, lv)
	lv.parents = append(lv.parents, vgn)

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

func (lg *LvmGraph) GetPhysicalVolume(name string) (*LvmNode, error) {
	id := fmt.Sprintf("pv:%s", name)
	node, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("physical volume %s does not exist", name)
	}
	return node, nil
}

func (lg *LvmGraph) GetVolumeGroup(name string) (*LvmNode, error) {
	id := fmt.Sprintf("vg:%s", name)
	node, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("volume group %s does not exist", name)
	}
	return node, nil
}

func (lg *LvmGraph) GetLogicalVolume(name string, vg string) (*LvmNode, error) {
	id := fmt.Sprintf("lv:%s:vg:%s", name, vg)
	node, found := lg.nodes[id]
	if !found {
		return nil, fmt.Errorf("logical volume %s/%s does not exist", vg, name)
	}
	return node, nil
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

func (lg *LvmGraph) Clear() {
	lg.nodes = map[string]*LvmNode{}
}
