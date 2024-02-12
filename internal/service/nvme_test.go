package service

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

const (
	UNSUPPORTED_NVME_VID = 0xFFFF
	UNSUPPORTED_NVME_MN  = "External NVME Manufacturer"
)

const (
	SpaceByte = 0x20
	NullByte  = 0x00
)

func TestGetBlockDeviceMapping(t *testing.T) {
	subtests := []struct {
		Name               string
		Device             string
		VendorId           uint16
		ModelNumber        string
		BlockDevice        string
		ExpectedOutput     string
		ExpectedError      error
	}{
		{
			Name:               "EBS NVMe Device + Pre-launch",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_EBS_MN,
			BlockDevice:        "sdb",
			ExpectedOutput:     "/dev/sdb",
			ExpectedError:      nil,
		},
		{
			Name:               "EBS NVMe Device + Post-launch",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_EBS_MN,
			BlockDevice:        "/dev/sdb",
			ExpectedOutput:     "/dev/sdb",
			ExpectedError:      nil,
		},
		{
			Name:               "Instance Store NVMe Device",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_INS_MN,
			BlockDevice:        "ephemeral0:sdb",
			ExpectedOutput:     "/dev/sdb",
			ExpectedError:      nil,
		},
		{
			Name:               "Instance Store NVMe Device + Null Byte",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_INS_MN,
			BlockDevice:        "ephemeral0:sdb\x00a",
			ExpectedOutput:     "/dev/sdb",
			ExpectedError:      nil,
		},
		{
			Name:               "Instance Store NVMe Device + Missing Block Device Mapping",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_INS_MN,
			BlockDevice:        "ephemeral0:none",
			ExpectedOutput:     "/dev/ephemeral0",
			ExpectedError:      nil,
		},
		{
			Name:               "Instance Store NVMe Device + Pattern Mismatch",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        AMZN_NVME_INS_MN,
			BlockDevice:        "ephemeral0:vdb",
			ExpectedOutput:     "",
			ExpectedError:      fmt.Errorf("🔴 /dev/nvme1n1: Instance-store vendor specific metadata did not match pattern. Pattern=^(ephemeral[0-9]):(sd[a-z]|none), Actual=ephemeral0:vdb"),
		},
		{
			Name:               "Invalid NVMe Device (Unsupported Vendor ID)",
			Device:             "/dev/nvme1n1",
			VendorId:           UNSUPPORTED_NVME_VID,
			ModelNumber:        AMZN_NVME_EBS_MN,
			BlockDevice:        "",
			ExpectedOutput:     "",
			ExpectedError:      fmt.Errorf("🔴 /dev/nvme1n1 is not an AWS-managed NVME device"),
		},
		{
			Name:               "Invalid NVMe Device (Unsupported Model Number)",
			Device:             "/dev/nvme1n1",
			VendorId:           AMZN_NVME_VID,
			ModelNumber:        UNSUPPORTED_NVME_MN,
			BlockDevice:        "",
			ExpectedOutput:     "",
			ExpectedError:      fmt.Errorf("🔴 /dev/nvme1n1 is not an AWS-managed NVME device"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			vsp, err := vendorSpecificPadding(subtest.ModelNumber)
			utils.ExpectErr("vendorSpecificPadding()", t, false, err)
			nd := &NVMeIoctlResult{
				Name: subtest.Device,
				IdCtrl: nvmeIdentifyController{
					Vid: subtest.VendorId,
					Mn:  modelNumber(subtest.ModelNumber, SpaceByte),
					Vs: nvmeIdentifyControllerAmznVS{
						Bdev: blockDevice(subtest.BlockDevice, vsp),
					},
				},
			}
			ns := NewAwsNitroNVMeService()
			bdm, err := ns.getBlockDeviceMapping(nd)
			utils.CheckError("getBlockDeviceMapping()", t, subtest.ExpectedError, err)
			utils.CheckOutput("getBlockDeviceMapping()", t, subtest.ExpectedOutput, bdm)
		})
	}
}

func modelNumber(input string, padding byte) [40]byte {
	var mn [40]byte
	// Copies input into mn[:]
	copy(mn[:], input)
	if len(input) < 40 {
		for i := len(input); i < 40; i++ {
			mn[i] = padding
		}
	}
	return mn
}

func blockDevice(input string, padding byte) [32]byte {
	var bd [32]byte
	// Copies input into bd[:]
	copy(bd[:], input)
	if len(input) < 32 {
		for i := len(input); i < 32; i++ {
			bd[i] = padding
		}
	}
	return bd
}

// The padding byte used for the Vendor Specific section depends on whether the NVMe Block
// device is an EBS (Space) or Instance Store (Null) Volume
func vendorSpecificPadding(modelNumber string) (byte, error) {
	switch modelNumber {
	case AMZN_NVME_EBS_MN:
		return SpaceByte, nil
	case AMZN_NVME_INS_MN:
		return NullByte, nil
	case UNSUPPORTED_NVME_MN:
		return NullByte, nil
	default:
		return NullByte, fmt.Errorf("🔴 %s: Could not determine vendor specific padding", modelNumber)
	}
}
