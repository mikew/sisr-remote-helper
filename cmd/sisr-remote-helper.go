package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	srh "github.com/mikew/sisr-remote-helper/internal"
)

const SISR_LAUCNH_DELAY = 7 * time.Second

func main() {
	manifest, err := srh.GetManifest()
	if err != nil {
		log.Fatalf("Failed to get manifest: %v", err)
	}

	logFile, prepareLoggerErr := srh.PrepareLogger()
	if prepareLoggerErr != nil {
		log.Fatalf("Failed to prepare logger: %v", prepareLoggerErr)
	}
	os.Stdout = logFile
	os.Stderr = logFile
	cleanup := func() {
		logFile.Close()
	}
	defer cleanup()

	cmd := &cli.Command{
		Name:    manifest.Name,
		Usage:   manifest.ShortDescription,
		Version: manifest.Version,

		Commands: []*cli.Command{
			&uwpCommand,
			&win32Command,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(fmt.Sprintf("Error running %s", manifest.Name), slog.Any("error", err))

		// Manually cleanup since defer won't run on os.Exit.
		cleanup()

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

		&cli.StringSliceFlag{
			Name:  "grep",
			Usage: "Also consider the app running if any process exe path contains this string (can be specified multiple times)",
		},
	},

	Action: func(ctx context.Context, cmd *cli.Command) error {
		aumid := cmd.Args().First()
		if aumid == "" {
			return fmt.Errorf("AUMID is required")
		}

		greps := cmd.StringSlice("grep")

		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
		done := make(chan error, 1)

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

			// Need to wait a bit before doing anything else.
			// Helps with vibration in Forza Horizon 5 at least, and TMNT Shredder's
			// Revenge seeing double controllers.
			time.Sleep(SISR_LAUCNH_DELAY)
		}

		defer func() {
			if sisrCmd != nil && sisrCmd.Process != nil {
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

		go func() {
			slog.Info("Launching app", slog.Any("aumid", aumid))
			// procAllowSetForeground.Call(uintptr(ASFW_ANY))
			done <- srh.StartAndWaitForUwpApp(aumid, greps)
		}()

		select {
		case <-ctx.Done():
			slog.Warn("Interrupted by user")

			targetFamily := strings.Split(aumid, "_")[0]
			srh.KillUwpApp(targetFamily, greps)

			return ctx.Err()
		case err := <-done:
			return err
		}
	},
}

var win32Command = cli.Command{
	Name:      "win32",
	Aliases:   []string{"exec"},
	Usage:     "Launch SISR and a Win32 executable",
	ArgsUsage: "<exe-path>",

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

		&cli.StringSliceFlag{
			Name:  "grep",
			Usage: "Also consider the app running if any process exe path contains this string (can be specified multiple times)",
		},
	},

	Action: func(ctx context.Context, cmd *cli.Command) error {
		exePath := cmd.Args().First()
		if exePath == "" {
			return fmt.Errorf("exe path is required")
		}

		greps := cmd.StringSlice("grep")
		exeArgs := cmd.Args().Tail()

		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
		done := make(chan error, 1)

		shouldStartSisr := cmd.Bool("start-sisr")

		var sisrCmd *exec.Cmd
		if shouldStartSisr {
			sisrCmdPath := cmd.String("sisr-path")
			sisrCmdArgs := []string{
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
			err := sisrCmd.Start()
			if err != nil {
				return fmt.Errorf("failed to start SISR: %w", err)
			}

			time.Sleep(SISR_LAUCNH_DELAY)
		}

		defer func() {
			if sisrCmd != nil && sisrCmd.Process != nil {
				slog.Info("Killing SISR helper", slog.Any("pid", sisrCmd.Process.Pid))
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

		go func() {
			slog.Info("Launching app", slog.Any("exe", exePath), slog.Any("args", exeArgs))
			done <- srh.StartAndWaitForWin32App(exePath, exeArgs, greps)
		}()

		select {
		case <-ctx.Done():
			slog.Warn("Interrupted by user")

			srh.KillWin32App(exePath, greps)

			return ctx.Err()
		case err := <-done:
			return err
		}
	},
}
