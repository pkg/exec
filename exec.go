// package exec is a universal wrapper around the os/exec package.
package exec

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

// System executes the command specified in command by calling /bin/sh -c command, and returns after the command has been completed. Stdin, Stdout, and Stderr are plumbed through to the child, but this behaviour can be modified by opts.
func System(command string, opts ...func(*Cmd) error) error {
	args := strings.Fields(command)
	args = append([]string{"-c"}, args...)
	cmd := Command("/bin/sh", args...)
	opts = append([]func(*Cmd) error{
		Stdin(os.Stdin),
		Stdout(os.Stdout),
		Stderr(os.Stderr),
	}, opts...)
	return cmd.Run(opts...)
}

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
	initalised    bool
	waited        bool
	before, after func(*Cmd) error
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
func (c *Cmd) Run(opts ...func(*Cmd) error) error {
	if err := c.Start(opts...); err != nil {
		return err
	}
	return c.Wait()
}

// Start starts the specified command but does not wait for it to complete.
//
// The Wait method will return the exit code and release associated resources
// once the command exits.
func (c *Cmd) Start(opts ...func(*Cmd) error) error {
	if !c.initalised {
		return errors.New("exec: command not initalised")
	}
	if err := applyDefaultOptions(c); err != nil {
		return err
	}
	if err := applyOptions(c, opts...); err != nil {
		return err
	}
	if c.before != nil {
		if err := c.before(c); err != nil {
			return err
		}
	}
	return c.Cmd.Start()
}

// Wait waits for the command to exit.
// It must have been started by Start.
func (c *Cmd) Wait() (err error) {
	if c.waited {
		return errors.New("exec: Wait was already called")
	}
	c.waited = true
	defer func() {
		if c.after == nil {
			return
		}
		errAfter := c.after(c)
		if err == nil {
			err = errAfter
		}
	}()
	return c.Cmd.Wait()
}

// Stdin specifies the process's standard input.
func Stdin(r io.Reader) func(*Cmd) error {
	return func(c *Cmd) error {
		if c.Stdin != nil {
			return errors.New("exec: Stdin already set")
		}
		c.Stdin = r
		return nil
	}
}

// Stdout specifies the process's standard output.
func Stdout(w io.Writer) func(*Cmd) error {
	return func(c *Cmd) error {
		if c.Stdout != nil {
			return errors.New("exec: Stdout already set")
		}
		c.Stdout = w
		return nil
	}
}

// Stderr specifies the process's standard error..
func Stderr(w io.Writer) func(*Cmd) error {
	return func(c *Cmd) error {
		if c.Stderr != nil {
			return errors.New("exec: Stderr already set")
		}
		c.Stderr = w
		return nil
	}
}

// BeforeFunc runs fn just prior to executing the command. If an error
// is returned, the command will not be run.
func BeforeFunc(fn func(*Cmd) error) func(*Cmd) error {
	return func(c *Cmd) error {
		if c.before != nil {
			return errors.New("exec: BeforeFunc already set")
		}
		c.before = fn
		return nil
	}
}

// AfterFunc runs fn just after to executing the command. If an error
// is returned, it will be returned providing the command exited cleanly.
func AfterFunc(fn func(*Cmd) error) func(*Cmd) error {
	return func(c *Cmd) error {
		if c.after != nil {
			return errors.New("exec: AfterFunc already set")
		}
		c.after = fn
		return nil
	}
}

// Setenv applies (or overwrites) childs environment key.
func Setenv(key, val string) func(*Cmd) error {
	return func(c *Cmd) error {
		key += "="
		for i := range c.Env {
			if strings.HasPrefix(c.Env[i], key) {
				c.Env[i] = key + val
				return nil
			}
		}
		c.Env = append(c.Env, key+val)
		return nil
	}
}

// Output runs the command and returns its standard output.
func (c *Cmd) Output(opts ...func(*Cmd) error) ([]byte, error) {
	var b bytes.Buffer
	opts = append([]func(*Cmd) error{Stdout(&b)}, opts...)
	err := c.Run(opts...)
	return b.Bytes(), err
}

// Dir specifies the working directory of the command.
// If Dir is empty, the command executes in the calling
// process's current directory.
func Dir(dir string) func(*Cmd) error {
	return func(c *Cmd) error {
		c.Dir = dir
		return nil
	}
}

func applyDefaultOptions(c *Cmd) error {
	if c.Env == nil {
		c.Env = os.Environ()
	}
	return nil
}

func applyOptions(c *Cmd, opts ...func(*Cmd) error) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}
