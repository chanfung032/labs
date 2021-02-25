package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

func usage() {
	fmt.Fprintln(os.Stderr, `
Usage:

  mint reload               # Reload Procfile
  mint start     PROC       # Start a Process
  mint stop      PROC       # Stop a Process
  mint restart   PROC       # Restart a Process
  mint status               # Print Processes' status
  mint check                # Show entries in Procfile
`)
	os.Exit(0)
}

// -- process information structure.
type procInfo struct {
	proc         string
	cmdline      string
	quit         bool
	cmd          *exec.Cmd
	mu           sync.Mutex
	cond         *sync.Cond
	waitErr      error
	pid          int
	statusTime   time.Time
	restartCount uint
}

// process informations named with proc.
var procs map[string]*procInfo

// filename of Procfile.
var procfile = flag.String("f", "/etc/Procfile", "proc file")

// filename of Mintenv.
var envfile = flag.String("e", "/etc/mintenv:.mintenv", "env file")

var logdir = flag.String("log_dir", "/var/log/mint/", "log directory")
var logMaxBytes = flag.Int("log_maxbytes", 10*1024*1024, "max log in bytes")
var logMaxBackups = flag.Int("log_maxbackups", 3, "max log backups to keep")

// rpc port number.
var port = flag.Uint("p", defaultPort(), "port")

// base directory
var basedir = flag.String("basedir", "", "base directory")

var maxProcNameLength = 0

// read Procfile and parse it.
func readProcfile(procfile string) (map[string]*procInfo, error) {
	procs := map[string]*procInfo{}
	content, err := ioutil.ReadFile(procfile)
	if err != nil {
		if os.IsNotExist(err) {
			return procs, nil
		}
		return procs, err
	}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}
		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		p := &procInfo{proc: k, cmdline: v}
		p.cond = sync.NewCond(&p.mu)
		procs[k] = p
		if len(k) > maxProcNameLength {
			maxProcNameLength = len(k)
		}
	}
	return procs, nil
}

// default port
func defaultPort() uint {
	s := os.Getenv("MINT_RPC_PORT")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err == nil {
			return uint(i)
		}
	}
	return 8555
}

// command: check. show Procfile entries.
func check() error {
	procs, err := readProcfile(*procfile)
	if err != nil {
		return err
	}
	keys := make([]string, len(procs))
	i := 0
	for k := range procs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	fmt.Printf("OK (%s)\n", strings.Join(keys, ", "))
	return nil
}

// command: start. spawn procs.
func master(args []string) error {
	var err error
	procs, err = readProcfile(*procfile)
	if err != nil {
		return err
	}
	if len(args) > 1 {
		tmp := map[string]*procInfo{}
		for _, v := range args[1:] {
			if _, exist := procs[v]; exist {
				tmp[v] = procs[v]
			} else {
				fmt.Fprintln(os.Stderr, "Unknown proc: "+v)
			}
		}
		procs = tmp
	}
	godotenv.Load(strings.Split(*envfile, ":")...)
	go func() {
		if err := startServer(); err != nil {
			panic(err)
		}
	}()
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, syscall.SIGCHLD)
	go func() {
		var sig os.Signal
		var status syscall.WaitStatus
		for {
			sig = <-sc
			if sig != syscall.SIGCHLD {
				break
			}
			syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		}
	}()
	err = startProcs()
	signal.Stop(sc)
	return err
}

func main() {
	var err error

	runtime.GOMAXPROCS(1)

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	if *basedir != "" {
		err = os.Chdir(*basedir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mint: %s\n", err.Error())
			os.Exit(1)
		}
	}

	cmd := flag.Args()[0]
	switch cmd {
	case "check":
		err = check()
		break
	case "start":
		fallthrough
	case "stop":
		fallthrough
	case "restart":
		if flag.NArg() == 2 {
			err = run(cmd, flag.Args()[1])
		} else {
			usage()
		}
		break
	case "status":
		fallthrough
	case "reload":
		if flag.NArg() == 1 {
			err = run(cmd, "")
		} else {
			usage()
		}
		break
	case "master":
		err = master(flag.Args())
		break
	case "version":
		fmt.Println(version)
		break
	default:
		usage()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
