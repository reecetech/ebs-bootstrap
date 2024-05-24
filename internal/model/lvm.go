package model

type LvmState int32
type LvmKind int32

const (
	DeviceActive             LvmState = 0b0000001
	PhysicalVolumeActive     LvmState = 0b0000010
	VolumeGroupInactive      LvmState = 0b0000100
	VolumeGroupActive        LvmState = 0b0001100
	LogicalVolumeInactive    LvmState = 0b0010000
	LogicalVolumeActive      LvmState = 0b0110000
	LogicalVolumeUnsupported LvmState = 0b1110000
)

const (
	DeviceKind         LvmKind = 0b0000001
	PhysicalVolumeKind LvmKind = 0b0000010
	VolumeGroupKind    LvmKind = 0b0000100
	LogicalVolumeKind  LvmKind = 0b0010000
)

type Device struct {
	Name string
	Size uint64
}

type PhysicalVolume struct {
	Name string
	Size uint64
}

type VolumeGroup struct {
	Name           string
	PhysicalVolume string
	Size           uint64
	State          LvmState
}

type LogicalVolume struct {
	Name        string
	VolumeGroup string
	State       LvmState
	Size        uint64
}
