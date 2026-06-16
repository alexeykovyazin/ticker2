//go:build windows

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tickerfile/internal/config"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/mgr"
)

func exePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}

func Install(cfg config.Config) error {
	path, err := exePath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.Service.Name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %q already exists", cfg.Service.Name)
	}

	serviceCfg := mgr.Config{
		DisplayName: cfg.Service.Name,
		Description: cfg.Service.Description,
		StartType:   mgr.StartAutomatic,
	}

	s, err = m.CreateService(cfg.Service.Name, path, serviceCfg)
	if err != nil {
		return err
	}
	defer s.Close()

	return nil
}

func Remove(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	return nil
}

func Start(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Start()
}

func Stop(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}

	timeout := 10
	for status.State != svc.Stopped {
		if timeout <= 0 {
			return fmt.Errorf("timed out waiting for service %q to stop", name)
		}
		timeout--
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return err
		}
	}
	return nil
}

func Run(cfg config.Config) error {
	return svc.Run(cfg.Service.Name, &handler{cfg: cfg})
}

func RunDebug(cfg config.Config) error {
	return debug.Run(cfg.Service.Name, &handler{cfg: cfg})
}
