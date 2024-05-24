package model

import "github.com/reecetech/ebs-bootstrap/internal/datastructures"

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
}

type LogicalVolumeState int32

const (
	Inactive    LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeInactive)
	Active      LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeActive)
	Unsupported LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeUnsupported)
)

type LogicalVolume struct {
	Name        string
	VolumeGroup string
	State       LogicalVolumeState
	Size        uint64
}
