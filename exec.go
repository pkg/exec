// package exec is a universal wrapper around the os/exec package.
package exec

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

// LookPath searches for an executable binary named file in the directories
// named by the PATH environment variable. If file contains a slash, it is
// tried directly and the PATH is not consulted. The result may be an
// absolute path or a path relative to the current directory.
func LookPath(file string) (string, error) { return exec.LookPath(file) }

// Command returns a Cmd to execute the named program with the given arguments.
func Command(name string, args ...string) *Cmd {
	return &Cmd{
		Cmd:        exec.Command(name, args...),
		initalised: true,
	}
}

// Cmd represents a command to be run.
// Cmd must be created by calling Command.
// Cmd cannot be reused after calling its Run or Start methods.
type Cmd struct {
	*exec.Cmd
	initalised bool
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *ExitError. Other error types may be
// returned for I/O problems.
func (c *Cmd) Run(opts ...func(*exec.Cmd) error) error {
	if err := c.Start(opts...); err != nil {
		return err
	}
	return c.Wait()
}

// Start starts the specified command but does not wait for it to complete.
//
// The Wait method will return the exit code and release associated resources
// once the command exits.
func (c *Cmd) Start(opts ...func(*exec.Cmd) error) error {
	if !c.initalised {
		return errors.New("exec: command not initalised")
	}
	if err := applyDefaultOptions(c.Cmd); err != nil {
		return err
	}
	if err := applyOptions(c.Cmd, opts...); err != nil {
		return err
	}
	return c.Cmd.Start()
}

// Stdin specifies the process's standard input.
func Stdin(r io.Reader) func(*exec.Cmd) error {
	return func(c *exec.Cmd) error {
		if c.Stdin != nil {
			return errors.New("exec: Stdin already set")
		}
		c.Stdin = r
		return nil
	}
}

// Stdout specifies the process's standard output.
func Stdout(w io.Writer) func(*exec.Cmd) error {
	return func(c *exec.Cmd) error {
		if c.Stdout != nil {
			return errors.New("exec: Stdout already set")
		}
		c.Stdout = w
		return nil
	}
}

// Stderr specifies the process's standard error..
func Stderr(w io.Writer) func(*exec.Cmd) error {
	return func(c *exec.Cmd) error {
		if c.Stderr != nil {
			return errors.New("exec: Stderr already set")
		}
		c.Stderr = w
		return nil
	}
}

// Output runs the command and returns its standard output.
func (c *Cmd) Output(opts ...func(*exec.Cmd) error) ([]byte, error) {
	var b bytes.Buffer
	opts = append([]func(*exec.Cmd) error{Stdout(&b)}, opts...)
	err := c.Run(opts...)
	return b.Bytes(), err
}

// Dir specifies the working directory of the command.
// If Dir is empty, the command executes in the calling
// process's current directory.
func Dir(dir string) func(*exec.Cmd) error {
	return func(c *exec.Cmd) error {
		c.Dir = dir
		return nil
	}
}

func applyDefaultOptions(c *exec.Cmd) error {
	return nil
}

func applyOptions(c *exec.Cmd, opts ...func(*exec.Cmd) error) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}
