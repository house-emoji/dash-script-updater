package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var repoDir string

func main() {
	if len(os.Args) < 2 {
		log.Println("the first argument must be a path to the repository to update")
		os.Exit(1)
	}

	repoDir = os.Args[1]

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		output, err := tryUpdate()
		if err != nil {
			http.Error(w, output, 500)
			log.Println("failed to update", err)
		} else {
			w.Write([]byte(output))
			log.Println("updated upon request")
		}
	})

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	log.Fatalln(http.ListenAndServe(":80", nil))
}

// tryUpdate attempts to update the Git repository and runs a post-update
// script. It returns the console output of the commands.
func tryUpdate() (string, error) {
	// Update the repository
	updateOutput, err := updateRepo()
	if err != nil {
		return updateOutput, err
	}

	// Run post-update script
	postUpdateOutput, err := postUpdate()
	if err != nil {
		return updateOutput + postUpdateOutput, err
	}

	return updateOutput + postUpdateOutput, nil
}

// updateRepo runs a git pull on the repository in the working directory. It
// returns the console output of the command.
func updateRepo() (string, error) {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrap(err, "running git pull")
	}

	return string(output), nil
}

// postUpdate runs the supplied post-update script. It returns the console
// output of the command.
func postUpdate() (string, error) {
	cmd := exec.Command("sh", "post-update.sh")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrap(err, "running post-update.sh")
	}

	return string(output), nil
}
