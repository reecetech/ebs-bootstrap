package main

import (
	"log"
	"os"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/layer"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func main() {
	log.SetFlags(0)

	// Services
	erf := utils.NewExecRunnerFactory()
	ufs := service.NewUnixFileService()
	lds := service.NewLinuxDeviceService(erf)
	uos := service.NewUnixOwnerService()
	ans := service.NewAwsNitroNVMeService()
	ls := service.NewLinuxLvmService(erf)
	fssf := service.NewLinuxFileSystemServiceFactory(erf)

	// Warnings
	warnings(uos)

	// Config + Flags
	c, err := config.New(os.Args)
	checkError(err)

	// Service + Config Consumers
	db := backend.NewLinuxDeviceBackend(lds, fssf)
	fb := backend.NewLinuxFileBackend(ufs)
	ub := backend.NewLinuxOwnerBackend(uos)
	dmb := backend.NewLinuxDeviceMetricsBackend(lds, fssf)
	lb := backend.NewLinuxLvmBackend(ls)

	// Executors
	dae := action.NewDefaultActionExecutor()
	le := layer.NewExponentialBackoffLayerExecutor(c, dae, layer.DefaultExponentialBackoffParameters())

	// Modify Config
	modifiers := []config.Modifier{
		config.NewAwsNVMeDriverModifier(ans, lds),
	}
	for _, m := range modifiers {
		checkError(m.Modify(c))
	}

	// LVM Layers
	lvmLayers := []layer.Layer{
		layer.NewCreatePhysicalVolumeLayer(db, lb),
		layer.NewCreateVolumeGroupLayer(lb),
	}
	checkError(le.Execute(lvmLayers))

	// LVM Modifiers
	lvmModifiers := []config.Modifier{
		config.NewLvmModifier(),
	}
	for _, m := range lvmModifiers {
		checkError(m.Modify(c))
	}

	// Validate Config
	validators := []config.Validator{
		config.NewFileSystemValidator(),
		config.NewModeValidator(),
		config.NewResizeThresholdValidator(),
		config.NewDeviceValidator(lds),
		config.NewMountPointValidator(),
		config.NewMountOptionsValidator(),
		config.NewOwnerValidator(uos),
	}
	for _, v := range validators {
		checkError(v.Validate(c))
	}

	// Layers
	layers := []layer.Layer{
		layer.NewFormatDeviceLayer(db),
		layer.NewLabelDeviceLayer(db),
		layer.NewCreateDirectoryLayer(fb),
		layer.NewMountDeviceLayer(db, fb),
		layer.NewResizeDeviceLayer(db, dmb),
		layer.NewChangeOwnerLayer(ub, fb),
		layer.NewChangePermissionsLayer(fb),
	}
	checkError(le.Execute(layers))
	log.Println("🟢 Passed all validation checks")
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func warnings(owns service.OwnerService) {
	user, err := owns.GetCurrentUser()
	if err != nil {
		return
	}
	if user.Id != 0 {
		log.Println("🚧 Not running as root user. Device operations might yield unexpected results")
	}
}
