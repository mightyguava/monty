package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/mightyguava/gomon/subproc"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gomon: ")
	log.SetOutput(os.Stderr)

	cmd := os.Args
	if len(cmd) < 2 {
		fmt.Println("Usage: gomon [command] [args ...]")
		os.Exit(1)
	}
	var args []string
	if len(cmd) > 2 {
		args = cmd[2:]
	}
	cmdString := strings.Join(cmd[1:], " ")
	executable := cmd[1]
	r := subproc.NewRunner(exec.Command(executable, args...))
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("error starting file watcher: ", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("error getting current directory: ", err)
	}
	err = w.Add(dir)
	if err != nil {
		log.Fatal("error starting file watcher: ", err)
	}
	log.Println("starting command: ", cmdString)
	err = r.Start()
	if err != nil {
		log.Fatal("error starting command: ", err.Error())
	}

	err = WatchAndRun(r, w)
	if err != nil {
		log.Fatal(err)
	}
}

// WatchAndRun watches for file changes and restarts the command. If it gets a SIGINT or SIGTERM, it
// will tell the child command to exit and then exit itself.
func WatchAndRun(r *subproc.Runner, w *fsnotify.Watcher) error {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-w.Events:
			log.Println("change detected, restarting command")
			if err := r.Restart(); err != nil {
				return err
			}
		case <-sigChan:
			log.Println("signal received, exiting")
			return r.Stop()
		}
	}
}
