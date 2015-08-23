// +build !windows,!plan9

package exec_test

import (
	"log"
	"os"

	"github.com/pkg/exec"
)

func ExampleCmd_Run() {
	cmd := exec.Command("git", "status")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCmd_Run_dir() {
	cmd := exec.Command("git", "status")
	// change working directory to /tmp and run git status.
	err := cmd.Run(exec.Dir("/tmp"))
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCmd_Run_stdout() {
	cmd := exec.Command("git", "status")
	// change working directory to /tmp, pass os.Stdout to the child
	// and run git status.
	err := cmd.Run(
		exec.Dir("/tmp"),
		exec.Stdout(os.Stdout),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCmd_Run_before() {
	cmd := exec.Command("/usr/sleep", "60s")
	// set a before and after function
	err := cmd.Run(
		exec.BeforeFunc(func(c *exec.Cmd) error {
			log.Println("About to call:", c.Args)
			return nil
		}),
		exec.AfterFunc(func(c *exec.Cmd) error {
			log.Println("Finished calling:", c.Args)
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCmd_Output_dir() {
	cmd := exec.Command("git", "status")
	// change working directory to /tmp, run git status and
	// capture the stdout of the child.
	out, err := cmd.Output(exec.Dir("/tmp"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", out)
}

func ExampleCmd_Start() {
	// start httpd server and redirect the childs stderr
	// to the parent's stdout.
	cmd := exec.Command("httpd")
	err := cmd.Start(exec.Stderr(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCmd_Start_Setenv() {
	// set or overwrite an environment variable.
	cmd := exec.Command("go")
	err := cmd.Start(exec.Setenv("GOPATH", "/foo"))
	if err != nil {
		log.Fatal(err)
	}
}
