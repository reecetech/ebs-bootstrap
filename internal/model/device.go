package model

import (
	"slices"
	"strings"
)

type BlockDevice struct {
	Name       string
	MountPoint string
	FileSystem FileSystem
	Label      string
}

type MountOptions string

func (mop MountOptions) Remount() MountOptions {
	mops := []string{}
	if len(mop) > 0 {
		mops = strings.Split(string(mop), ",")
	}

	// If not found. index == -1
	index := slices.Index(mops, "remount")
	if index < 0 {
		mops = append(mops, "remount")
	}
	return MountOptions(strings.Join(mops, ","))
}

type BlockDeviceMetrics struct {
	FileSystemSize  uint64
	BlockDeviceSize uint64
}

func (bdm *BlockDeviceMetrics) ShouldResize(threshold float64) bool {
	// Minimum File System Size (mfss)
	mfss := float64(bdm.BlockDeviceSize) * (threshold / 100)
	return float64(bdm.FileSystemSize) < mfss
}
