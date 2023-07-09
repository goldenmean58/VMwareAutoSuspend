package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/execabs"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type handler struct {
	elog debug.Log
}

func (h *handler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptPreShutdown | svc.AcceptStop | svc.AcceptShutdown}

loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
			// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
			time.Sleep(100 * time.Millisecond)
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			break loop
		case svc.PreShutdown:
			h.suspendAllVMs()
		default:
			h.elog.Warning(1, fmt.Sprintf("unexpected control request %#v", c))
		}
	}
	s <- svc.Status{State: svc.StopPending}
	return
}

const vmRunPath = `C:\Program Files (x86)\VMware\VMware Workstation\vmrun.exe`

func (h *handler) suspendAllVMs() {
	h.elog.Info(1, "Start suspending all VMs")
	vmList, err := vmRunList()
	if err != nil {
		h.elog.Error(1, fmt.Sprintf("Failed to list VMs: %v", err))
		return
	}
	for _, vm := range vmList {
		h.elog.Info(1, fmt.Sprintf("Suspend VM %s", vm))
		if err := vmRunSuspend(vm); err != nil {
			h.elog.Error(1, fmt.Sprintf("Failed to suspend VM %s: %v", vm, err))
		}
	}
}

func vmRun(args ...string) (output []byte, err error) {
	cmd := execabs.Command(vmRunPath, args...)
	return cmd.CombinedOutput()
}

func vmRunList() ([]string, error) {
	output, err := vmRun("list")
	if err != nil {
		return nil, err
	}
	var l []string
	for i, scanner := 0, bufio.NewScanner(bytes.NewBuffer(output)); scanner.Scan(); i++ {
		if i == 0 {
			scanner.Bytes()
			continue
		}
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			l = append(l, line)
		}
	}
	return l, nil
}

func vmRunSuspend(vm string) error {
	_, err := vmRun("suspend", vm)
	// 正常执行也不会输出任何内容
	return err
}
