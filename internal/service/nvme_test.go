package service

import (
	"fmt"
	"testing"
	"ebs-bootstrap/internal/utils"
)

const (
	UNSUPPORTED_NVME_VID = 0xFFFF
	UNSUPPORTED_NVME_MN	 = "External NVME Manufacturer"
)

func TestAwsNVMeService(t *testing.T) {
	subtests := []struct{
		Name 			string
		Device			string
		VendorId		uint16
		ModelNumber		string
		BlockDevice		string
		ExpectedOutput	string
		ExpectedErr		error
	}{
		{
			Name:			"EBS NVMe Device (Partial Block Device)",
			Device:			"/dev/nvme1n1",
			VendorId:		AMZN_NVME_VID,
			ModelNumber:	AMZN_NVME_EBS_MN,
			BlockDevice:	"sdb",
			ExpectedOutput:	"/dev/sdb",
			ExpectedErr:	nil,
		},
		{
			Name:			"EBS NVMe Device (Complete Block Device)",
			Device:			"/dev/nvme1n1",
			VendorId:		AMZN_NVME_VID,
			ModelNumber:	AMZN_NVME_EBS_MN,
			BlockDevice:	"/dev/sdb",
			ExpectedOutput:	"/dev/sdb",
			ExpectedErr:	nil,
		},
		{
			Name:			"Invalid NVMe Device (Unsupported Vendor ID)",
			Device:			"/dev/nvme1n1",
			VendorId:		UNSUPPORTED_NVME_VID,
			ModelNumber:	AMZN_NVME_EBS_MN,
			BlockDevice:	"",
			ExpectedOutput:	"",
			ExpectedErr:	fmt.Errorf("🔴 /dev/nvme1n1 is not an AWS-managed NVME device"),
		},
		{
			Name:			"Invalid NVMe Device (Unsupported Model Number)",
			Device:			"/dev/nvme1n1",
			VendorId:		AMZN_NVME_VID,
			ModelNumber:	UNSUPPORTED_NVME_MN,
			BlockDevice:	"",
			ExpectedOutput:	"",
			ExpectedErr:	fmt.Errorf("🔴 /dev/nvme1n1 is not an AWS-managed NVME device"),
		},
	}
	for _, subtest := range subtests {
        t.Run(subtest.Name, func(t *testing.T) {
			nd := &NVMeDevice{
				Name:	subtest.Device,
				IdCtrl:	nvmeIdentifyController{
					Vid:	subtest.VendorId,
					Mn:		parseModelNumber(subtest.ModelNumber),
					Vs:		nvmeIdentifyControllerAmznVS{
						Bdev:	parseBlockDevice(subtest.BlockDevice),
					},
				},
			}
			ns := &AwsNVMeService{}
			bdm, err := ns.getBlockDeviceMapping(nd)
			if bdm != subtest.ExpectedOutput {
				t.Errorf("getBlockDeviceMapping() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, bdm)
			}
            utils.CheckError("getBlockDeviceMapping()", t, subtest.ExpectedErr, err)
        })
    }
}

func parseModelNumber(input string) [40]byte {
	var mn [40]byte
	copy(mn[:], input)
	if len(input) < 40 {
		for i := len(input); i < 40; i++ {
			mn[i] = ' '
		}
	}
	return mn
}

func parseBlockDevice(input string) [32]byte {
	var bd [32]byte
	copy(bd[:], input)
	if len(input) < 32 {
		for i := len(input); i < 32; i++ {
			bd[i] = ' '
		}
	}
	return bd
}
