package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Executor struct {
	Force, Log, Done bool
	Errors           []error
}

func NewExecutor() Executor {
	return Executor{Force: false, Log: false, Done: false, Errors: make([]error, 0)}
}

func (e *Executor) Execute(command string, args ...string) (output string) {
	if e.Done {
		return ""
	}
	for index, value := range args {
		command = strings.Replace(command, fmt.Sprintf("$%d", index+1), value, -1)
	}
	if e.Log {
		fmt.Println("Executor running: " + command)
	}
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if e.Force {
			e.Done = true
		}
		e.Errors = append(e.Errors, err)
	}
	if e.Log {
		fmt.Println(string(out))
	}
	return string(out)
}

func (e Executor) DidError() (errored bool) {
	return len(e.Errors) > 0 || e.Done
}

func (e Executor) FormatErrors() (output string) {
	output = ""
	for _, err := range e.Errors {
		if output == "" {
			output += err.Error()
		} else {
			output += "\n" + err.Error()
		}
	}
	return ""
}
