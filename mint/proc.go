package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	var (
		err          error
		restartCount uint = 0
		resetTimer   *time.Timer
		cmd          *exec.Cmd
	)
	logger := &Logger{
		Filename:   fmt.Sprintf("%s/%s.log", *logdir, proc),
		MaxBackups: *logMaxBackups,
		MaxBytes:   int64(*logMaxBytes),
	}

	cs := []string{"/bin/sh", "-c", "exec " + procs[proc].cmdline}
	fmt.Fprintf(logger, "Starting %s\n", proc)
	resetTimer = time.AfterFunc(time.Minute, func() {
		restartCount = 0
	})
	procs[proc].restartCount = 0
restart:
	cmd = exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return true
	}
	procs[proc].pid = cmd.Process.Pid
	procs[proc].statusTime = time.Now()
	procs[proc].cmd = cmd
	procs[proc].mu.Unlock()
	err = cmd.Wait()
	procs[proc].mu.Lock()
	procs[proc].cond.Broadcast()
	procs[proc].waitErr = err
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)
	if restartCount < 3 && procs[proc].quit == false {
		procs[proc].restartCount++
		restartCount++
		resetTimer.Reset(time.Minute)
		fmt.Fprintf(logger, "Restarting %s\n", proc)
		cmd.Process = nil
		cmd.ProcessState = nil
		goto restart
	}
	resetTimer.Stop()
	procs[proc].statusTime = time.Now()

	return procs[proc].quit
}

func killPidGroup(pid int, sig syscall.Signal) error {
	// use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	pid = -1 * pid
	target, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return target.Signal(sig)
}

// stop specified proc.
func stopProc(proc string) error {
	p, ok := procs[proc]
	if !ok {
		return errors.New("Unknown proc: " + proc)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	p.quit = true
	err := killPidGroup(p.pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	timer := time.AfterFunc(5*time.Second, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p, ok := procs[proc]; ok && p.cmd != nil && p.cmd.Process != nil {
			err = killPidGroup(p.cmd.Process.Pid, syscall.SIGKILL)
		}
	})
	p.cond.Wait()
	timer.Stop()
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	p, ok := procs[proc]
	if !ok {
		return errors.New("Unknown proc: " + proc)
	}

	p.mu.Lock()
	if procs[proc].cmd != nil {
		p.mu.Unlock()
		return nil
	}

	wg.Add(1)
	go func() {
		spawnProc(proc)
		wg.Done()
		p.mu.Unlock()
	}()
	return nil
}

// restart specified proc.
func restartProc(proc string) error {
	if _, ok := procs[proc]; !ok {
		return errors.New("Unknown proc: " + proc)
	}
	stopProc(proc)
	return startProc(proc)
}

// spawn all procs.
func startProcs() error {
	for proc := range procs {
		startProc(proc)
	}
	sc := make(chan os.Signal, 10)
	/* //terminate master if no child process exist
	go func() {
		wg.Wait()
		sc <- syscall.SIGINT
	}()
	*/
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-sc
	for proc := range procs {
		stopProc(proc)
	}
	wg.Wait() // wait all child exit
	return nil
}

func reloadProcs(newProcs map[string]*procInfo) error {
	for proc := range procs {
		if err := stopProc(proc); err != nil {
			return err
		}
	}
	procs = newProcs
	for proc := range procs {
		startProc(proc)
	}
	return nil
}
