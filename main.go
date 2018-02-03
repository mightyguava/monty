package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Runner manages running, stopping, and restarting the command we want to execute
type Runner struct {
	templateCmd *exec.Cmd
	runningCmd  *exec.Cmd
	done        chan error

	mu      *sync.Mutex
	running bool
}

// NewRunner allocates and initializes a new Runner
func NewRunner(cmd *exec.Cmd) *Runner {
	return &Runner{
		templateCmd: cmd,
		mu:          &sync.Mutex{},
	}
}

// Start executes the command in a separate process
func (r *Runner) Start() error {
	c := &exec.Cmd{}
	*c = *r.templateCmd
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r.runningCmd = c
	r.done = make(chan error)
	err := c.Start()
	if err != nil {
		return err
	}
	r.running = true
	go func() {
		err := c.Wait()
		log.Println("command exited")

		r.mu.Lock()
		r.running = false
		r.mu.Unlock()

		r.done <- err
	}()
	return nil
}

// Stop signals the command to stop with SIGINT and waits for it to stop. If it does not stop after
// 3 seconds, it kills the command
func (r *Runner) Stop() error {
	c := r.runningCmd
	if c == nil {
		return nil
	}
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	log.Println("sending command SIGINT")
	// Sending a signal to the negative PID signals the whole process group. Runner.Start() starts the
	// child process within a process group.
	if err := syscall.Kill(-c.Process.Pid, syscall.SIGINT); err != nil {
		return err
	}
	select {
	case <-r.done:
		return nil
	case <-time.After(3 * time.Second):
		if err := syscall.Kill(-c.Process.Pid, syscall.SIGKILL); err != nil {
			return errors.New("failed to kill: " + err.Error())
		}
		log.Println("timed out waiting for command to stop. KILLED")
	}
	return nil
}

// Restart calls Stop and Start in sequence
func (r *Runner) Restart() error {
	err := r.Stop()
	if err != nil {
		return err
	}
	return r.Start()
}

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
	r := NewRunner(exec.Command(executable, args...))
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("error starting file watcher: ", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("error getting current directly: ", err)
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
func WatchAndRun(r *Runner, w *fsnotify.Watcher) error {
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
