package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	sish "github.com/mikew/sish/internal"
	"github.com/urfave/cli/v3"
)

func main() {
	manifest, err := sish.GetManifest()
	if err != nil {
		log.Fatalf("Failed to get manifest: %v", err)
	}

	logFile, prepareLoggerErr := sish.PrepareLogger()
	if prepareLoggerErr != nil {
		log.Fatalf("Failed to prepare logger: %v", prepareLoggerErr)
	}
	os.Stdout = logFile
	os.Stderr = logFile
	defer logFile.Close()

	cmd := &cli.Command{
		Name:    manifest.Name,
		Usage:   manifest.ShortDescription,
		Version: manifest.Version,

		Commands: []*cli.Command{
			&uwpCommand,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(fmt.Sprintf("Error running %s", manifest.Name), slog.Any("error", err))
		os.Exit(1)
	}
}

var uwpCommand = cli.Command{
	Name:      "uwp",
	Usage:     "Launch SISR and a UWP app",
	ArgsUsage: "<aumid>",

	Flags: []cli.Flag{
		&cli.BoolWithInverseFlag{
			Name:  "start-sisr",
			Usage: "Whether to start SISR automatically",
			Value: true,
		},

		&cli.StringFlag{
			Name:  "sisr-path",
			Value: "./SISR",
		},

		&cli.StringFlag{
			Name: "sisr-config",
		},
	},

	Action: func(ctx context.Context, cmd *cli.Command) error {
		aumid := cmd.Args().Get(0)
		if aumid == "" {
			return fmt.Errorf("AUMID is required")
		}

		shouldStartSisr := cmd.Bool("start-sisr")

		var sisrCmd *exec.Cmd
		if shouldStartSisr {
			sisrCmdPath := cmd.String("sisr-path")
			sisrCmdArgs := []string{
				// "--tray", "false",

				// "--window-create", "false",
				// "--window-fullscreen", "true",
				// "--window-continous-draw", "true",

				// "--gyro-passthrough",

				"--marker",
				"--debug",
				"--log-level", "debug",
				"--log-file", "sisr-remote-helper-sisr.log",
			}

			if config := cmd.String("sisr-config"); config != "" {
				sisrCmdArgs = append(sisrCmdArgs, "--config", config)
			}

			slog.Info("Starting SISR", slog.Any("command", sisrCmdPath), slog.Any("args", sisrCmdArgs))
			sisrCmd = exec.Command(sisrCmdPath, sisrCmdArgs...)
			sisrCmd.Stdout = os.Stdout
			sisrCmd.Stderr = os.Stderr
			// Somehow this makes window focus MUCH more annoying.
			// sisrCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			err := sisrCmd.Start()
			if err != nil {
				return fmt.Errorf("failed to start SISR: %w", err)
			}
		}

		defer func() {
			if shouldStartSisr && sisrCmd != nil && sisrCmd.Process != nil {
				slog.Info("Killing SISR helper", slog.Any("pid", sisrCmd.Process.Pid))
				// sisrCmd.Process.Kill()
				killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", sisrCmd.Process.Pid))
				killCmd.Stdout = os.Stdout
				killCmd.Stderr = os.Stderr
				killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				if err := killCmd.Run(); err != nil {
					slog.Warn("Failed to kill SISR process", slog.Any("error", err))
				}

				time.Sleep(1 * time.Second)

				slog.Info("Running steam://forceinputappid/0")
				steamForceInputCmd := exec.Command("cmd", "/c", "start", "steam://forceinputappid/0")
				steamForceInputCmd.Stdout = os.Stdout
				steamForceInputCmd.Stderr = os.Stderr
				steamForceInputCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				if err := steamForceInputCmd.Run(); err != nil {
					slog.Warn("Failed to send forceinputappid command", slog.Any("error", err))
				}
			}
		}()

		slog.Info("Launching app", slog.Any("aumid", aumid))
		// procAllowSetForeground.Call(uintptr(ASFW_ANY))
		if err := sish.StartAndWaitForUwpApp(aumid); err != nil {
			return fmt.Errorf("failed to start and wait for UWP app: %w", err)
		}

		return nil
	},
}
