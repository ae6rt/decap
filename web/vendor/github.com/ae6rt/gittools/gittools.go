package gittools

import (
	"log"
	"os"
	"os/exec"
)

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

// Clone clones the Git repository at repositoryURL at the given branch into the given directory.  If shallow is true, use --depth 1.
func Clone(repositoryURL, branch, dir string, shallow bool) error {
	args := []string{"clone"}
	if shallow {
		args = append(args, []string{"--depth", "1"}...)
	}
	args = append(args, []string{"--branch", branch, repositoryURL, dir}...)
	return executeShellCommand("git", args)
}

func executeShellCommand(commandName string, args []string) error {
	logger.Printf("Executing %s %+v\n", commandName, args)
	command := exec.Command(commandName, args...)
	var stdOutErr []byte
	var err error
	stdOutErr, err = command.CombinedOutput()
	if err != nil {
		return err
	}
	logger.Printf("%v\n", string(stdOutErr))

	return nil
}
