package service

import (
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type LvmService interface {
	CreatePhysicalVolume(name string) error
}

type LinuxLvmService struct {
	runnerFactory utils.RunnerFactory
}

type PvsResponse struct {
	Report []struct {
		PhysicalVolume []struct {
			Name string `json:"pv_name"`
		} `json:"pv"`
	} `json:"report"`
}

func NewLinuxLvmService(rf utils.RunnerFactory) *LinuxLvmService {
	return &LinuxLvmService{
		runnerFactory: rf,
	}
}

func (ls *LinuxLvmService) CreatePhysicalVolume(name string) error {
	r := ls.runnerFactory.Select(utils.PvCreate)
	_, err := r.Command(name)
	return err
}
