//go:build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tickerfile/internal/service"

	"golang.org/x/sys/windows/svc"
)

var (
	svcName = "tickerfile"
	logDir  = ""
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command> [flags]\n"+
			" where <command> is one of\n"+
			" install, remove, debug, start, stop\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	flag.StringVar(&svcName, "name", svcName, "name of the Windows service")
	flag.StringVar(&logDir, "logdir", "", "directory for log files (default: directory of executable)")
	flag.Parse()

	if logDir == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("failed to determine executable path: %v", err)
		}
		logDir = filepath.Dir(exe)
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if running in service: %v", err)
	}
	if inService {
		if err := service.Run(svcName, logDir); err != nil {
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
		if err := service.RunDebug(svcName, logDir); err != nil {
			log.Fatalf("debug service failed: %v", err)
		}
	case "install":
		if err := service.Install(svcName, "Writes timestamps to log files every 2 seconds"); err != nil {
			log.Fatalf("install failed: %v", err)
		}
		fmt.Printf("service %q installed\n", svcName)
	case "remove":
		if err := service.Remove(svcName); err != nil {
			log.Fatalf("remove failed: %v", err)
		}
		fmt.Printf("service %q removed\n", svcName)
	case "start":
		if err := service.Start(svcName); err != nil {
			log.Fatalf("start failed: %v", err)
		}
		fmt.Printf("service %q started\n", svcName)
	case "stop":
		if err := service.Stop(svcName); err != nil {
			log.Fatalf("stop failed: %v", err)
		}
		fmt.Printf("service %q stopped\n", svcName)
	default:
		usage(fmt.Sprintf("invalid command %q", cmd))
	}
}
