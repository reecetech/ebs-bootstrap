package datastructures

import (
	"fmt"
)

type LvmNodeState int32
type LvmNodeCategory int32

const (
	BlockDeviceActive        LvmNodeState = 0b0000001
	PhysicalVolumeActive     LvmNodeState = 0b0000010
	VolumeGroupInactive      LvmNodeState = 0b0000100
	VolumeGroupActive        LvmNodeState = 0b0001100
	LogicalVolumeInactive    LvmNodeState = 0b0010000
	LogicalVolumeActive      LvmNodeState = 0b0110000
	LogicalVolumeUnsupported LvmNodeState = 0b1110000
)

const (
	BlockDevice    LvmNodeCategory = 0b0000001
	PhysicalVolume LvmNodeCategory = 0b0000010
	VolumeGroup    LvmNodeCategory = 0b0000100
	LogicalVolume  LvmNodeCategory = 0b0010000
)

type LvmNode struct {
	id       string
	Name     string
	State    LvmNodeState
	children []*LvmNode
	parents  []*LvmNode
}

func NewBlockDevice(name string) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("device:%s", name),
		Name:     name,
		State:    BlockDeviceActive,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewPhysicalVolume(name string) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("pv:%s", name),
		Name:     name,
		State:    PhysicalVolumeActive,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewVolumeGroup(name string) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("vg:%s", name),
		Name:     name,
		State:    VolumeGroupInactive,
		children: []*LvmNode{},
		parents:  []*LvmNode{},
	}
}

func NewLogicalVolume(name string, vg string, State LvmNodeState) *LvmNode {
	return &LvmNode{
		id:       fmt.Sprintf("lv:%s:vg:%s", name, vg),
		Name:     name,
		State:    State,
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

func (lg *LvmGraph) AddBlockDevice(name string) error {
	bd := NewBlockDevice(name)

	_, found := lg.nodes[bd.id]
	if found {
		return fmt.Errorf("block device %s already exists", name)
	}

	lg.nodes[bd.id] = bd
	return nil
}

func (lg *LvmGraph) AddPhysicalVolume(name string) error {
	pv := NewPhysicalVolume(name)

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

func (lg *LvmGraph) AddVolumeGroup(name string, pv string) error {
	id := fmt.Sprintf("vg:%s", name)

	vg, found := lg.nodes[id]
	if !found {
		vg = NewVolumeGroup(name)
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

func (lg *LvmGraph) AddLogicalVolume(name string, vg string, state LvmNodeState) error {
	lv := NewLogicalVolume(name, vg, state)

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
		if lvn.State == LogicalVolumeActive {
			vgn.State = VolumeGroupActive
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

func (lg *LvmGraph) GetParents(node *LvmNode, state LvmNodeCategory) []*LvmNode {
	parents := []*LvmNode{}
	for _, p := range node.parents {
		// Bitmasking to check if the parent nodes is of the desired category
		if int32(p.State)&int32(state) > 0 {
			parents = append(parents, p)
		}
	}
	return parents
}

func (lg *LvmGraph) GetChildren(node *LvmNode, state LvmNodeCategory) []*LvmNode {
	children := []*LvmNode{}
	for _, c := range node.children {
		// Bitmasking to check if children nodes is of the desired category
		if int32(c.State)&int32(state) > 0 {
			children = append(children, c)
		}
	}
	return children
}
