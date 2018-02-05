package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mightyguava/monty/livereload"
	"github.com/mightyguava/monty/subproc"
	"github.com/rjeczalik/notify"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("monty: ")
	log.SetOutput(os.Stderr)

	urlFlag := flag.String("url", "", "a URL to open in the browser and to live reload")
	flag.Parse()
	fmt.Println(*urlFlag)

	cmd := flag.Args()
	if len(cmd) == 0 && *urlFlag == "" {
		fmt.Println("Usage: monty [flags ...] [command] [args ...]")
		os.Exit(1)
	}

	w, err := CreateWatcher()
	if err != nil {
		log.Fatal("error starting file watcher: ", err)
	}

	var r *subproc.Runner
	if len(cmd) > 0 {
		var args []string
		if len(cmd) > 1 {
			args = cmd[1:]
		}
		executable := cmd[0]
		r = subproc.NewRunner(exec.Command(executable, args...))
		if err = r.Start(); err != nil {
			log.Fatal("error starting command: ", err.Error())
		}
	}

	var chrome *livereload.Chrome
	if *urlFlag != "" {
		if chrome, err = livereload.NewChrome(*urlFlag); err != nil {
			r.Stop()
			log.Fatal("could not connect to chrome: ", err)
		}
		log.Println("opening Chrome to:", *urlFlag)
		if err = chrome.Open(); err != nil {
			r.Stop()
			log.Fatal("could not open url: ", *urlFlag)
		}
	}

	reloader := NewReloader(r, chrome, w)
	if err = reloader.WatchAndRun(); err != nil {
		reloader.Shutdown()
		log.Fatal(err)
	}
}

// CreateWatcher creates and returns a fs watcher for the current working directory
func CreateWatcher() (chan notify.EventInfo, error) {
	c := make(chan notify.EventInfo, 100)
	if err := notify.Watch("./...", c, notify.All); err != nil {
		return nil, err
	}
	return c, nil
}
