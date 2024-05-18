package config

import (
	"fmt"
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

type LvmModifier struct{}

func NewLvmModifier() *LvmModifier {
	return &LvmModifier{}
}

func (lm *LvmModifier) Modify(c *Config) error {
	// Fetch a copy of the original keys as we are updating
	// the config in-place and it is unsafe to iterate over it
	// directly
	keys := make([]string, len(c.Devices))
	for name := range c.Devices {
		keys = append(keys, name)
	}
	for _, key := range keys {
		device := c.Devices[key]
		if len(device.Lvm) > 0 {
			ldn := fmt.Sprintf("/dev/%s/%s", device.Lvm, device.Lvm)
			c.Devices[ldn] = device
			delete(c.Devices, key)
		}
	}
	return nil
}
