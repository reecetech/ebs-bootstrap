package service

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type FileSystemService interface {
	GetSize(name string) (uint64, error)
	GetFileSystem() model.FileSystem
	Format(name string) error
	Label(name string, label string) error
	Resize(name string) error
	GetMaximumLabelLength() int
	DoesResizeRequireMount() bool
	DoesLabelRequireUnmount() bool
}

type FileSystemServiceFactory interface {
	Select(fs model.FileSystem) (FileSystemService, error)
}

type LinuxFileSystemServiceFactory struct {
	RunnerFactory utils.RunnerFactory
}

func NewLinuxFileSystemServiceFactory(rc utils.RunnerFactory) *LinuxFileSystemServiceFactory {
	return &LinuxFileSystemServiceFactory{
		RunnerFactory: rc,
	}
}

func (fsf *LinuxFileSystemServiceFactory) Select(fs model.FileSystem) (FileSystemService, error) {
	switch fs {
	case model.Ext4:
		return NewExt4Service(fsf.RunnerFactory), nil
	case model.Xfs:
		return NewXfsService(fsf.RunnerFactory), nil
	case model.Unformatted:
		//lint:ignore ST1005 This Error Message Is Supposed to Be Prepended With A Device Name
		return nil, fmt.Errorf("An unformatted file system can not be queried/modified")
	default:
		//lint:ignore ST1005 This Error Message Is Supposed to Be Prepended With A Device Name
		return nil, fmt.Errorf("Support for querying/modifying the '%s' filesystem is lacking", fs.String())
	}
}

type Ext4Service struct {
	runnerFactory utils.RunnerFactory
}

func NewExt4Service(rc utils.RunnerFactory) *Ext4Service {
	return &Ext4Service{runnerFactory: rc}
}

func (es *Ext4Service) GetFileSystem() model.FileSystem {
	return model.Ext4
}

func (es *Ext4Service) Format(name string) error {
	r := es.runnerFactory.Select(utils.MkfsExt4)
	_, err := r.Command(name)
	return err
}

func (es *Ext4Service) Label(name string, label string) error {
	r := es.runnerFactory.Select(utils.E2Label)
	_, err := r.Command(name, label)
	return err
}

func (es *Ext4Service) Resize(name string) error {
	r := es.runnerFactory.Select(utils.Resize2fs)
	_, err := r.Command(name)
	return err
}

func (es *Ext4Service) GetSize(name string) (uint64, error) {
	r := es.runnerFactory.Select(utils.Tune2fs)
	output, err := r.Command("-l", name)
	if err != nil {
		return 0, err
	}
	// Regex (Block Size)
	rebs := regexp.MustCompile(`Block size:\s+(\d+)`)
	// Match (Block Size)
	mbs := rebs.FindStringSubmatch(output)
	if len(mbs) != 2 {
		return 0, fmt.Errorf("🔴 %s: Block size not found tune2fs output", name)
	}
	// String (Block Size)
	sbs := mbs[1]
	// Block Size
	bs, err := strconv.ParseUint(sbs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("🔴 Failed to cast block size to unsigned 64-bit integer")
	}

	// Regex (Block Count)
	rebc := regexp.MustCompile(`Block count:\s+(\d+)`)
	// Match (Block Count)
	mbc := rebc.FindStringSubmatch(output)
	if len(mbs) != 2 {
		return 0, fmt.Errorf("🔴 %s: Block count not found tune2fs output", name)
	}
	// String (Block Count)
	sbc := mbc[1]
	// Block Count
	bc, err := strconv.ParseUint(sbc, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("🔴 Failed to cast block size to unsigned 64-bit integer")
	}
	return bs * bc, nil
}

func (es *Ext4Service) GetMaximumLabelLength() int {
	return 16
}

func (es *Ext4Service) DoesResizeRequireMount() bool {
	return false
}

func (es *Ext4Service) DoesLabelRequireUnmount() bool {
	return false
}

type XfsService struct {
	runnerFactory utils.RunnerFactory
}

func NewXfsService(rc utils.RunnerFactory) *XfsService {
	return &XfsService{runnerFactory: rc}
}

func (es *XfsService) GetFileSystem() model.FileSystem {
	return model.Xfs
}

func (xs *XfsService) Format(name string) error {
	r := xs.runnerFactory.Select(utils.MkfsXfs)
	_, err := r.Command(name)
	return err
}

func (xs *XfsService) Label(name string, label string) error {
	r := xs.runnerFactory.Select(utils.XfsAdmin)
	_, err := r.Command("-L", label, name)
	return err
}

func (es *XfsService) Resize(name string) error {
	r := es.runnerFactory.Select(utils.XfsGrowfs)
	_, err := r.Command(name)
	return err
}

func (xs *XfsService) GetSize(name string) (uint64, error) {
	r := xs.runnerFactory.Select(utils.XfsInfo)
	output, err := r.Command(name)
	if err != nil {
		return 0, err
	}
	// Regex (Data)
	red := regexp.MustCompile(`data\s+=\s+bsize=(\d+)\s+blocks=(\d+)`)
	// Match (Data)
	md := red.FindStringSubmatch(output)
	if len(md) != 3 {
		return 0, fmt.Errorf("🔴 %s: Block size and block count not found xfs_info output", name)
	}
	// String (Block Size)
	sbs := md[1]
	// Block Size
	bs, err := strconv.ParseUint(sbs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("🔴 Failed to cast block size to unsigned 64-bit integer")
	}
	// String (Block Count)
	sbc := md[2]
	// Block Count
	bc, err := strconv.ParseUint(sbc, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("🔴 Failed to cast block count to unsigned 64-bit integer")
	}
	return bs * bc, nil
}

func (es *XfsService) GetMaximumLabelLength() int {
	return 12
}

func (es *XfsService) DoesResizeRequireMount() bool {
	return true
}

func (es *XfsService) DoesLabelRequireUnmount() bool {
	return true
}
