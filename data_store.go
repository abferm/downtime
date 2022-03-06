package downtime

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DataStore interface {
	SetStamp(t time.Time) error
	GetStamp() (time.Time, error)
	SetShutdown(t time.Time) error
	GetShutdown() (time.Time, error)
	SetBoot(t time.Time) error
	GetBoot() (time.Time, error)
}

func NewDataDir(dir string) (*DataDir, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to load datadir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return &DataDir{dir}, nil
}

type DataDir struct {
	dir string
}

func (dd DataDir) stampFile() string {
	return filepath.Join(dd.dir, "downtimed.stamp")
}

func (dd DataDir) SetStamp(t time.Time) error {
	return touch(dd.stampFile(), t)
}

func (dd DataDir) GetStamp() (time.Time, error) {
	return stat(dd.stampFile())
}

func (dd DataDir) shutdownFile() string {
	return filepath.Join(dd.dir, "downtimed.shutdown")
}

func (dd DataDir) SetShutdown(t time.Time) error {
	return touch(dd.shutdownFile(), t)
}

func (dd DataDir) GetShutdown() (time.Time, error) {
	return stat(dd.shutdownFile())
}

func (dd DataDir) bootFile() string {
	return filepath.Join(dd.dir, "downtimed.boot")
}

func (dd DataDir) SetBoot(t time.Time) error {
	return touch(dd.bootFile(), t)
}

func (dd DataDir) GetBoot() (time.Time, error) {
	return stat(dd.bootFile())
}

func touch(name string, t time.Time) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()
	err = os.Chtimes(name, t, t)
	return err
}

func stat(name string) (time.Time, error) {
	info, err := os.Stat(name)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
