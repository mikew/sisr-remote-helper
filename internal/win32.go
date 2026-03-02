package sisr_remote_helper

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func FindPidsMatchingGreps(greps []string) []int32 {
	if len(greps) == 0 {
		return nil
	}

	foundPids := []int32{}
	pids, _ := process.Pids()

	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		exePath, err := proc.Exe()
		if err != nil {
			continue
		}

		for _, grep := range greps {
			if strings.Contains(strings.ToLower(exePath), strings.ToLower(grep)) {
				foundPids = append(foundPids, pid)
			}
		}
	}

	return foundPids
}

func StartAndWaitForWin32App(exePath string, args []string, greps []string) error {
	win32Cmd := exec.Command(exePath, args...)
	win32Cmd.Stdout = os.Stdout
	win32Cmd.Stderr = os.Stderr

	if err := win32Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	slog.Info("Waiting for app to close", slog.Any("pid", win32Cmd.Process.Pid))

	mainExited := make(chan struct{})
	go func() {
		win32Cmd.Wait()
		close(mainExited)
	}()

	mainProcessRunning := true
	didGiveGracePeriod := false

	for {
		if mainProcessRunning {
			select {
			case <-mainExited:
				mainProcessRunning = false
				slog.Info("Main process exited")
			default:
			}
		}

		isRunning := mainProcessRunning

		if !isRunning && len(greps) > 0 {
			if !didGiveGracePeriod {
				slog.Info("Main process exited, waiting for grep matches to appear")
				time.Sleep(5 * time.Second)
				didGiveGracePeriod = true
			}

			isRunning = len(FindPidsMatchingGreps(greps)) > 0
		}

		if !isRunning {
			slog.Info("Win32 app closed")
			break
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func KillWin32App(exePath string, greps []string) {
	// Kill by exe path first, then any grep-matched processes.
	// Using FindPidsMatchingGreps with the exe path works because it does a
	// substring match against each process's full exe path.
	allGreps := append([]string{exePath}, greps...)
	for _, pid := range FindPidsMatchingGreps(allGreps) {
		exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid)).Run()
	}
}
