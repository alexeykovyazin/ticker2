//go:build windows

package service

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"tickerfile/internal/logwriter"

	"golang.org/x/sys/windows/svc"
)

type handler struct {
	logDir string
}

func (h *handler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	s <- svc.Status{State: svc.StartPending}

	if err := os.MkdirAll(h.logDir, 0o755); err != nil {
		log.Printf("failed to create log directory %q: %v", h.logDir, err)
		return false, 1
	}

	textLog, err := logwriter.OpenText(filepath.Join(h.logDir, "text.log"))
	if err != nil {
		log.Printf("failed to open text log: %v", err)
		return false, 1
	}
	defer textLog.Close()

	win32Log, err := logwriter.OpenWin32(filepath.Join(h.logDir, "win32.log"))
	if err != nil {
		log.Printf("failed to open win32 log: %v", err)
		return false, 1
	}
	defer win32Log.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case <-ticker.C:
			ts := time.Now().UTC().Format(time.RFC3339)
			line := ts + "\n"

			if err := textLog.WriteLine(line); err != nil {
				log.Printf("text log write failed: %v", err)
			}
			if err := win32Log.AppendLine(line); err != nil {
				log.Printf("win32 log write failed: %v", err)
			}

		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				return false, 0
			default:
				log.Printf("unexpected service control request: %d", c.Cmd)
			}
		}
	}
}
