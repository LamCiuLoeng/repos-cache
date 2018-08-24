package git

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

// GitCommand is a command to be executed by git
type GitCommand struct {
	ProcInput *bytes.Reader
	Args      []string
}

// Run runs the git command
func (gitCommand *GitCommand) Run(wait bool) (io.ReadCloser, error) {
	log.Printf("Executing: %v", gitCommand.Args)
	// cmd := exec.Command("git", gitCommand.Args...)
	cmd := exec.Command(gitCommand.Args[0], gitCommand.Args[1:]...)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Println("err 1")
		return nil, err
	}

	if gitCommand.ProcInput != nil {
		cmd.Stdin = gitCommand.ProcInput
	}

	if err := cmd.Start(); err != nil {
		log.Println("err 2")
		return nil, err
	}

	if wait {
		err = cmd.Wait()
		if err != nil {
			log.Println("err 3")
			return nil, err
		}
	}

	return stdout, nil
}

// RunAndGetOutput runs the command and gets the output
func (gitCommand *GitCommand) RunAndGetOutput() []byte {
	stdout, err := gitCommand.Run(false)
	if err != nil {
		return []byte{}
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		return []byte{}
	}

	return data
}

// WriteGitToHTTP copies the output of the git command to the http socket.
func (gitCommand *GitCommand) WriteGitToHTTP(w http.ResponseWriter, wait bool) {
	stdout, err := gitCommand.Run(wait)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("run command error:", err)
		return
	}

	nbytes, err := io.Copy(w, stdout)

	if err != nil {
		log.Fatal("Error writing to socket", err)
	} else {
		log.Printf("Bytes written: %d", nbytes)
	}
}
