//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	svcName = "VMwareAutoSuspender"
)

func usage(errmsg string) {
	if errmsg != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", errmsg)
	}
	fmt.Fprintf(os.Stderr,
		"Usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove", os.Args[0])
	os.Exit(2)
}

func main() {
	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		runService(false)
		return
	}
	if len(os.Args) < 2 {
		usage("no command specified")
	}
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "install":
		err = installService()
	case "remove":
		err = removeService()
	case "debug":
		runService(true)
	}
	if err != nil {
		log.Fatalf("failed to %s: %v", cmd, err)
	}
}

func runService(isDebug bool) {
	var (
		err  error
		elog debug.Log
	)
	if isDebug {
		elog = debug.New(svcName)
	} else {
		elog, err = eventlog.Open(svcName)
		if err != nil {
			log.Fatalf("failed to open event log: %v", err)
		}
	}
	defer elog.Close()

	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(svcName, &handler{elog: elog})
	if err != nil {
		elog.Error(1, fmt.Sprintf("run service failed: %v", err))
		return
	}
}
