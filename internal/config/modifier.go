package config

import (
	"log"
	"strings"

	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type Modifier interface {
	Modify(c *Config) error
}

type AwsNitroNVMeModifier struct {
	nvmeService   service.NVMeService
	deviceService service.DeviceService
}

func NewAwsNVMeDriverModifier(nvmeService service.NVMeService, deviceService service.DeviceService) *AwsNitroNVMeModifier {
	return &AwsNitroNVMeModifier{
		nvmeService:   nvmeService,
		deviceService: deviceService,
	}
}

func (andm *AwsNitroNVMeModifier) Modify(c *Config) error {
	bds, err := andm.deviceService.GetBlockDevices()
	if err != nil {
		return err
	}
	for _, name := range bds {
		// Check if device already exists in the config
		// No need to make additional queries if this is the case
		_, exists := c.Devices[name]
		if exists {
			continue
		}
		if !strings.HasPrefix(name, "/dev/nvme") {
			continue
		}
		bdm, err := andm.nvmeService.GetBlockDeviceMapping(name)
		if err != nil {
			return err
		}
		log.Printf("🔵 Nitro NVMe detected: %s -> %s", name, bdm)
		cd, exists := c.Devices[bdm]
		// We can detect AWS NVMe Devices, but this doesn't neccesarily
		// mean they will be managed through configuration
		if !exists {
			continue
		}
		// Delete the original reference to the device configuration from the
		// block device mapping retrieved from the NVMe IoCtl interface and
		// replace it with the actual device name
		//	Before:
		// 		/dev/sdb => *config.Device (a)
		//	After:
		//		/dev/nvme0n1 => *config.Device (a)
		c.Devices[name] = cd
		delete(c.Devices, bdm)
	}
	return nil
}
