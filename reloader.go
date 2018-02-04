package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mightyguava/monty/livereload"
	"github.com/mightyguava/monty/subproc"
	"github.com/rjeczalik/notify"
	"golang.org/x/time/rate"
)

// Reloader handles restarting the command and reloading chrome
type Reloader struct {
	r      *subproc.Runner
	chrome *livereload.Chrome
	w      chan notify.EventInfo
}

// NewReloader initializes and returns a new Reloader instance
func NewReloader(r *subproc.Runner, chrome *livereload.Chrome, w chan notify.EventInfo) *Reloader {
	return &Reloader{
		r:      r,
		chrome: chrome,
		w:      w,
	}
}

// WatchAndRun watches for file changes and restarts the command. If it gets a SIGINT or SIGTERM, it
// will tell the child command to exit and then exit itself.
func (r *Reloader) WatchAndRun() error {
	var err error

	limitWindow := 500 * time.Millisecond
	mu := &sync.Mutex{}
	reloadPending := false

	lm := rate.NewLimiter(rate.Every(limitWindow), 1)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error)
	for {
		select {
		case <-r.w:
			mu.Lock()
			if !reloadPending {
				reloadPending = true
				go func() {
					lm.Wait(context.Background())
					if err := r.Reload(); err != nil {
						errChan <- err
					}
					reloadPending = false
				}()
			}
			mu.Unlock()
		case <-sigChan:
			log.Println("signal received, exiting")
			if r.r != nil {
				if err = r.r.Stop(); err != nil {
					log.Println("error stopping process: ", err)
				}
			}
			if r.chrome != nil {
				if err = r.chrome.Close(); err != nil {
					log.Println("error closing Chrome: ", err)
				}
			}
			return nil
		}
	}
}

// Reload starts the command or reloads the chrome window
func (r *Reloader) Reload() error {
	if r.r != nil {
		log.Println("change detected, restarting command")
		if err := r.r.Restart(); err != nil {
			return err
		}
	}
	if r.chrome != nil {
		log.Println("change detected, reloading chrome")
		if err := r.chrome.Reload(); err != nil {
			return err
		}
	}
	return nil
}
