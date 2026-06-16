//go:build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tickerfile/internal/config"
	"tickerfile/internal/service"

	"golang.org/x/sys/windows/svc"
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command> [flags]\n"+
			" where <command> is one of\n"+
			" install, remove, debug, start, stop, init-config\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func loadConfig(configPath string) config.Config {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to determine executable path: %v", err)
	}
	exeDir := filepath.Dir(exe)

	cfg := config.Default(exeDir)
	path := config.ResolvePath(exe, configPath)
	if _, err := os.Stat(path); err == nil {
		loaded, err := config.Load(path)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		cfg = loaded
	}
	cfg.ApplyDefaults(exeDir)
	return cfg
}

func main() {
	var (
		configPath = ""
		svcName    = ""
		logDir     = ""
	)

	flag.StringVar(&configPath, "config", "", "path to configuration file (default: tickerfile.json next to executable)")
	flag.StringVar(&svcName, "name", "", "override service name from config")
	flag.StringVar(&logDir, "logdir", "", "override log directory from config")
	flag.Parse()

	cfg := loadConfig(configPath)
	if svcName != "" {
		cfg.Service.Name = svcName
	}
	if logDir != "" {
		cfg.Log.Dir = logDir
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if running in service: %v", err)
	}
	if inService {
		if err := service.Run(cfg); err != nil {
			log.Fatalf("service failed: %v", err)
		}
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		usage("no command specified")
	}
	cmd := strings.ToLower(args[0])

	switch cmd {
	case "debug":
		if err := service.RunDebug(cfg); err != nil {
			log.Fatalf("debug service failed: %v", err)
		}
	case "init-config":
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("failed to determine executable path: %v", err)
		}
		path := config.ResolvePath(exe, configPath)
		if err := config.WriteDefault(path, filepath.Dir(exe)); err != nil {
			log.Fatalf("init-config failed: %v", err)
		}
		fmt.Printf("wrote configuration to %s\n", path)
	case "install":
		if err := service.Install(cfg); err != nil {
			log.Fatalf("install failed: %v", err)
		}
		fmt.Printf("service %q installed\n", cfg.Service.Name)
	case "remove":
		if err := service.Remove(cfg.Service.Name); err != nil {
			log.Fatalf("remove failed: %v", err)
		}
		fmt.Printf("service %q removed\n", cfg.Service.Name)
	case "start":
		if err := service.Start(cfg.Service.Name); err != nil {
			log.Fatalf("start failed: %v", err)
		}
		fmt.Printf("service %q started\n", cfg.Service.Name)
	case "stop":
		if err := service.Stop(cfg.Service.Name); err != nil {
			log.Fatalf("stop failed: %v", err)
		}
		fmt.Printf("service %q stopped\n", cfg.Service.Name)
	default:
		usage(fmt.Sprintf("invalid command %q", cmd))
	}
}
