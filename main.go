package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const (
	runEvery = time.Second * 10

	groupName   = "tabhouse"
	repoName    = "dash-scripts"
	serviceName = "amazon-dash"

	lastUpdateFileName = ".dash-script-last-update.json"
)

type lastUpdate struct {
	SHA1 string `json:"sha1"`
}

func main() {
	for range time.Tick(runEvery) {
		updated, err := tryUpdate()
		if err != nil {
			log.Println("update attempt failed", err)
		} else if updated {
			log.Println("updated repostory")
		} else {
			log.Println("update was not necessary")
		}
	}
}

// tryUpdate attempts to update the Git repository. It returns true if it had
// to update.
func tryUpdate() (bool, error) {
	lu, err := loadLastUpdateFile()

	client := github.NewClient(nil)

	// Load all recent commits
	listOpts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page: 1}}
	commits, _, err := client.Repositories.ListCommits(context.Background(), groupName, repoName, listOpts)
	if err != nil {
		return false, err
	}

	// Compare the most recent commit to our most recent update
	if commits[0].GetSHA() != lu.SHA1 {
		// Update the repository
		err = updateRepo()
		if err != nil {
			return false, errors.Wrap(err, "updating "+serviceName)
		}

		// Restart the service
		err = restartService()
		if err != nil {
			return false, errors.Wrap(err, "restarting "+serviceName)
		}

		// Save this last update info
		err = saveLastUpdateFile(lastUpdate{
			SHA1: commits[0].GetSHA()})
		if err != nil {
			return false, errors.Wrap(err, "saving last update file")
		}

		return true, nil
	}

	return false, nil
}

// loadLastUpdateFile loads info on the last update from a file.
func loadLastUpdateFile() (lastUpdate, error) {
	var output lastUpdate

	home, err := homeDir()
	if err != nil {
		return lastUpdate{}, err
	}

	// Load the last update file
	file, err := os.Open(filepath.Join(home, lastUpdateFileName))
	if err == os.ErrNotExist {
		// The file doesn't exist yet, so we assume the repo needs updating
		return lastUpdate{SHA1: ""}, nil
	} else {
		return lastUpdate{}, errors.Wrap(err, "opening last update file")
	}
	defer file.Close()

	// Decode it as a JSON file
	err = json.NewDecoder(file).Decode(&output)
	if err != nil {
		return lastUpdate{}, errors.Wrap(err, "decoding last update file")
	}

	return output, nil
}

// saveLastUpdateFile saves the given last update data to a file.
func saveLastUpdateFile(lu lastUpdate) error {
	home, err := homeDir()
	if err != nil {
		return err
	}

	// Create the last update file
	file, err := os.Create(filepath.Join(home, lastUpdateFileName))
	if err != nil {
		return errors.Wrap(err, "creating last update file")
	}
	defer file.Close()

	// Encode the data as JSON
	err = json.NewEncoder(file).Encode(lu)
	if err != nil {
		return errors.Wrap(err, "encoding last update file")
	}

	return nil
}

// updateRepo runs a git pull on the repository in the working directory.
func updateRepo() error {
	cmd := exec.Command("git", "pull")
	err := cmd.Run()
	if err != nil {
		if cmd.Stdin != nil {
			// Load the output of the failing command
			output, ioErr := ioutil.ReadAll(cmd.Stdin)
			if ioErr != nil {
				errors.Wrap(ioErr, "loading command output after error '"+err.Error()+"'")
				return ioErr
			}
			return errors.Wrap(errors.New(string(output)), "running git pull")
		}
		return errors.Wrap(err, "running git pull")
	}

	return nil
}

// restartService restarts the Systemd service specified by serviceName.
func restartService() error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	err := cmd.Run()
	if err != nil {
		if cmd.Stdin != nil {
			// Load the output of the failing command
			output, ioErr := ioutil.ReadAll(cmd.Stdin)
			if ioErr != nil {
				errors.Wrap(ioErr, "loading command output after error '"+err.Error()+"'")
				return ioErr
			}
			return errors.Wrap(errors.New(string(output)), "running systemctl restart "+serviceName)
		}
		return errors.Wrap(err, "running systemctl restart "+serviceName)
	}

	return nil
}

// homeDir returns the home directory of the current user.
func homeDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "looking up home directory")
	}

	return user.HomeDir, nil
}
