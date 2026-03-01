package sish

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	ProcessQueryLimitedInfo = 0x1000
)

const (
	ASFW_ANY = 0xFFFFFFFF
	// SW_SHOW    = 5
	// SW_RESTORE = 9
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess  = kernel32.NewProc("OpenProcess")
	procGetPkgFamily = kernel32.NewProc("GetPackageFamilyName")
	procCloseHandle  = kernel32.NewProc("CloseHandle")
	// user32           = syscall.NewLazyDLL("user32.dll")
	// procAllowSetForeground = user32.NewProc("AllowSetForegroundWindow")

	// procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	// procShowWindow               = user32.NewProc("ShowWindow")
	// procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	// procEnumWindows              = user32.NewProc("EnumWindows")
)

func GetPackageFamilyName(pid int32) (string, error) {
	handle, _, _ := procOpenProcess.Call(uintptr(ProcessQueryLimitedInfo), 0, uintptr(pid))
	if handle == 0 {
		return "", fmt.Errorf("could not open process")
	}
	defer procCloseHandle.Call(handle)

	var length uint32 = 256
	buffer := make([]uint16, length)

	// Returns 0 on success (ERROR_SUCCESS)
	ret, _, _ := procGetPkgFamily.Call(handle, uintptr(unsafe.Pointer(&length)), uintptr(unsafe.Pointer(&buffer[0])))
	if ret != 0 {
		return "", fmt.Errorf("not a UWP app")
	}

	return syscall.UTF16ToString(buffer[:length]), nil
}

func FindPidsForFamily(family string) []int32 {
	foundPids := []int32{}

	pids, _ := process.Pids()
	for _, pid := range pids {
		f, err := GetPackageFamilyName(pid)

		if err == nil && strings.Contains(strings.ToLower(f), strings.ToLower(family)) {
			foundPids = append(foundPids, pid)
		}
	}

	return foundPids
}

func StartAndWaitForUwpApp(aumid string) error {
	targetFamily := strings.Split(aumid, "_")[0]

	// exec.Command("explorer", `shell:AppsFolder\`+aumid).Run()
	uwpCmd := exec.Command("cmd", "/c", "start", `shell:AppsFolder\`+aumid)
	uwpCmd.Stdout = os.Stdout
	uwpCmd.Stderr = os.Stderr
	uwpCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := uwpCmd.Run(); err != nil {
		return err
	}

	slog.Info("Waiting for app to close")
	time.Sleep(10 * time.Second)

	// focused := false
	didGiveGracePeriod := false

	for {
		isRunning := len(FindPidsForFamily(targetFamily)) > 0

		if isRunning {
			// If we haven't successfully grabbed focus yet, try now
			// if !focused {
			// 	if forceFocus(targetFamily) {
			// 		slog.Info("Focus successfully handed to UWP app.")
			// 		focused = true
			// 	}
			// }
		} else {
			// Grace period to ensure it's not just a splash-screen handoff
			if !didGiveGracePeriod {
				time.Sleep(5 * time.Second)
				didGiveGracePeriod = true
			}

			if len(FindPidsForFamily(targetFamily)) < 1 {
				slog.Info("UWP App closed")
				break
			}
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

// func forceFocus(family string) bool {
// 	pids, _ := process.Pids()
// 	foundWindow := false

// 	for _, pid := range pids {
// 		f, err := GetPackageFamilyName(pid)
// 		if err == nil && strings.Contains(strings.ToLower(f), strings.ToLower(family)) {
// 			// Find the HWND for this PID
// 			procEnumWindows.Call(syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
// 				var windowPid uint32
// 				procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&windowPid)))

// 				if windowPid == uint32(pid) {
// 					// Check if the window is actually visible/functional
// 					// 9 = SW_RESTORE, 5 = SW_SHOW
// 					procAllowSetForeground.Call(uintptr(pid))
// 					procShowWindow.Call(hwnd, SW_RESTORE)
// 					procSetForegroundWindow.Call(hwnd)

// 					foundWindow = true
// 					return 0 // Stop EnumWindows
// 				}
// 				return 1 // Keep looking
// 			}), 0)
// 		}
// 		if foundWindow {
// 			break
// 		}
// 	}
// 	return foundWindow
// }
