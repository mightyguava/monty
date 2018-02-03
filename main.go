package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mightyguava/gomon/livereload"
	"github.com/mightyguava/gomon/subproc"
	"github.com/rjeczalik/notify"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gomon: ")
	log.SetOutput(os.Stderr)

	urlFlag := flag.String("url", "", "a URL to open in the browser and to live reload")
	flag.Parse()
	fmt.Println(*urlFlag)

	cmd := flag.Args()
	if len(cmd) == 0 && *urlFlag == "" {
		fmt.Println("Usage: gomon [command] [args ...]")
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
		cmdString := strings.Join(cmd, " ")
		executable := cmd[0]
		r = subproc.NewRunner(exec.Command(executable, args...))
		log.Println("starting command: ", cmdString)
		if err = r.Start(); err != nil {
			log.Fatal("error starting command: ", err.Error())
		}
	}

	var chrome *livereload.Chrome
	if *urlFlag != "" {
		url, err := url.Parse(*urlFlag)
		if err != nil {
			log.Fatal("invalid url: ", *urlFlag)
		}
		if url.Scheme == "" {
			url.Scheme = "http"
		}
		if chrome, err = livereload.NewChrome(url.String()); err != nil {
			log.Fatal("could not connect to chrome: ", err)
		}
		log.Println("opening Chrome to: ", url.String())
		if err = chrome.Open(); err != nil {
			log.Fatal("could not open url: ", url.String())
		}
	}

	if err = WatchAndRun(r, chrome, w); err != nil {
		log.Fatal(err)
	}
}

// CreateWatcher creates and returns a fs watcher for the current working directory
func CreateWatcher() (chan notify.EventInfo, error) {
	c := make(chan notify.EventInfo)
	if err := notify.Watch("./...", c, notify.All); err != nil {
		return nil, err
	}
	return c, nil
}

// WatchAndRun watches for file changes and restarts the command. If it gets a SIGINT or SIGTERM, it
// will tell the child command to exit and then exit itself.
func WatchAndRun(r *subproc.Runner, chrome *livereload.Chrome, w chan notify.EventInfo) error {
	var err error

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-w:
			if r != nil {
				log.Println("change detected, restarting command")
				if err = r.Restart(); err != nil {
					return err
				}
			}
			if chrome != nil {
				log.Println("change detected, reloading chrome")
				if err = chrome.Reload(); err != nil {
					return err
				}
			}
		case <-sigChan:
			log.Println("signal received, exiting")
			if r != nil {
				if err = r.Stop(); err != nil {
					log.Println("error stopping process: ", err)
				}
			}
			if chrome != nil {
				if err = chrome.Close(); err != nil {
					log.Println("error closing Chrome: ", err)
				}
			}
			return nil
		}
	}
}
