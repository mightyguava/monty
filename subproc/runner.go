package subproc

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Runner manages running, stopping, and restarting the command we want to execute
type Runner struct {
	templateCmd *exec.Cmd
	runningCmd  *exec.Cmd
	cmdString   string
	done        chan error

	mu      *sync.Mutex
	running bool
}

// NewRunner allocates and initializes a new Runner
func NewRunner(cmd *exec.Cmd) *Runner {
	return &Runner{
		templateCmd: cmd,
		cmdString:   strings.Join(cmd.Args, " "),
		mu:          &sync.Mutex{},
	}
}

// Start executes the command in a separate process
func (r *Runner) Start() error {
	log.Println("starting command:", r.cmdString)

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
			return errors.New("failed to kill:" + err.Error())
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
