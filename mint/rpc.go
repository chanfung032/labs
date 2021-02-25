package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Mint int

// rpc: start
func (r *Mint) Start(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	newProcs, err := readProcfile(*procfile)
	if err != nil {
		return err
	}
	if _, exist := newProcs[proc]; exist {
		if _, exist := procs[proc]; exist {
			procs[proc].cmdline = newProcs[proc].cmdline
		} else {
			procs[proc] = newProcs[proc]
		}
	}
	return startProc(proc)
}

// rpc: stop
func (r *Mint) Stop(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return stopProc(proc)
}

// rpc: restart
func (r *Mint) Restart(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	newProcs, err := readProcfile(*procfile)
	if err != nil {
		return err
	}
	if _, exist := newProcs[proc]; exist {
		if _, exist := procs[proc]; exist {
			procs[proc].cmdline = newProcs[proc].cmdline
		} else {
			procs[proc] = newProcs[proc]
		}
	}
	return restartProc(proc)
}

func getUptimeString(sec time.Duration) string {
	day := sec / time.Second / (60 * 60 * 24)
	hour := sec / time.Second / (60 * 60) % 24
	minute := sec / time.Second / 60 % 60
	second := sec / time.Second % 60
	ret := fmt.Sprintf("%d:%02d:%02d", hour, minute, second)
	if day > 0 {
		s := ""
		if day > 1 {
			s = "s"
		}
		ret = fmt.Sprintf("%d day%s, %s", day, s, ret)
	}
	return ret
}

// rpc: status
func (r *Mint) Status(empty string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	keys := []string{}
	for proc := range procs {
		keys = append(keys, proc)
	}
	sort.Strings(keys)
	for _, proc := range keys {
		namefmt := fmt.Sprintf("%%-%ds", maxProcNameLength)
		if procs[proc].cmd != nil {
			*ret += fmt.Sprintf(namefmt+"    RUNNING    pid %d, uptime %s", proc, procs[proc].pid, getUptimeString(time.Since(procs[proc].statusTime)))
		} else if !procs[proc].quit && procs[proc].waitErr != nil {
			*ret += fmt.Sprintf(namefmt+"    EXITED     %s, %s", proc, procs[proc].waitErr.Error(), procs[proc].statusTime.Local().Format("Jan 2 03:04 PM"))
		} else {
			*ret += fmt.Sprintf(namefmt+"    STOPPED    %s", proc, procs[proc].statusTime.Local().Format("Jan 2 03:04 PM"))
		}
		if procs[proc].restartCount > 0 {
			s := ""
			if procs[proc].restartCount > 1 {
				s = "s"
			}
			*ret += fmt.Sprintf(", restarted %d time%s", procs[proc].restartCount, s)
		}
		*ret += "\n"
	}
	return err
}

func (r *Mint) Reload(empty string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	// start new proc, and stop old ones
	newProcs, err := readProcfile(*procfile)
	if err != nil {
		return err
	}
	godotenv.Overload(strings.Split(*envfile, ":")...)
	return reloadProcs(newProcs)
}

// command: run.
func run(cmd, arg string) error {
	client, err := rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return err
	}
	defer client.Close()
	var ret string
	switch cmd {
	case "start":
		return client.Call("Mint.Start", arg, &ret)
	case "stop":
		return client.Call("Mint.Stop", arg, &ret)
	case "restart":
		return client.Call("Mint.Restart", arg, &ret)
	case "reload":
		return client.Call("Mint.Reload", arg, &ret)
	case "status":
		err := client.Call("Mint.Status", arg, &ret)
		fmt.Print(ret)
		return err
	}
	return errors.New("Unknown command " + cmd)
}

// start rpc server.
func startServer() error {
	gm := new(Mint)
	rpc.Register(gm)
	server, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return err
	}
	for {
		client, err := server.Accept()
		if err != nil {
			continue
		}
		rpc.ServeConn(client)
	}
	return nil
}
