package main

import (
	"os"
	"log"
	"ebs-bootstrap/internal/config"
	"ebs-bootstrap/internal/service"
	"ebs-bootstrap/internal/utils"
	"ebs-bootstrap/internal/state"
)

func main() {
	// Disable Timetamp
	log.SetFlags(0)
	e := utils.NewExecRunner()
	ds := &service.LinuxDeviceService{Runner: e}
	ns := &service.AwsNVMeService{}
	fs := &service.UnixFileService{}
	dts := &service.EbsDeviceTranslator{DeviceService: ds, NVMeService: ns}

	dt, err := dts.GetTranslator()
	if err != nil {
		log.Fatal(err)
	}
	config, err := config.New(os.Args, dt, fs)
	if err != nil {
		log.Fatal(err)
	}

	for name, device := range config.Devices {
		d, err := state.NewDevice(name, ds, fs)
		if err != nil {
			log.Fatal(err)
		}
		err = d.Diff(config)
		if err == nil {
			log.Printf("🟢 %s: No changes detected", name)
			continue
		}
		if device.Mode == "healthcheck" {
			log.Fatal(err)
		}
	}
}
