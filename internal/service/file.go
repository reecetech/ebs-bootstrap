package service

import (
	"os"
	"syscall"
	"fmt"
)

// File Service Interface [START]

type FileInfo struct {
	Owner		string
	Group		string
	Permissions	string
	Exists		bool
}

type FileService interface {
	GetStats(file string) (*FileInfo, error)
	ValidateFile(path string)	(error)
}

// File Service Interface [END]

type UnixFileService struct {}

func (ds *UnixFileService) GetStats(file string) (*FileInfo, error) {
	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileInfo{Exists: false}, nil
		}
        return nil, err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return &FileInfo{
			Owner: fmt.Sprintf("%d", stat.Uid),
			Group: fmt.Sprintf("%d", stat.Gid),
			Permissions: fmt.Sprintf("%o", info.Mode().Perm()),
			Exists:	true,
		}, nil
	}
	return nil, fmt.Errorf("🔴 %s: Failed to get stats", file)
}

func (ds *UnixFileService) ValidateFile(path string) (error) {
    s, err := os.Stat(path)
    if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("🔴 %s does not exist", path)
		}
        return err
    }
    if !s.Mode().IsRegular() {
        return fmt.Errorf("🔴 %s is not a regular file", path)
    }
    return nil
}

func (ds *UnixFileService) ValidateDirectory(path string) (error) {
    s, err := os.Stat(path)
    if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("🔴 %s does not exist", path)
		}
        return err
    }
    if !s.Mode().IsDir() {
        return fmt.Errorf("🔴 %s is not a directory", path)
    }
    return nil
}
