package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type UpgradeCommand struct {
	Name string
	Args []string
}

type UpgradeSequence []UpgradeCommand

func buildDebianUpgrade(package_list map[string]string) UpgradeSequence {
	updateCmd := "apt-get"
	updateArg := []string{"update", "-q"}

	installCmd := "apt-get"
	installArg := []string{"install", "--only-upgrade", "--no-install-recommends", "-y", "-q"}

	for name, version := range package_list {
		installArg = append(installArg, name+"="+version)
	}

	return UpgradeSequence{UpgradeCommand{updateCmd, updateArg}, UpgradeCommand{installCmd, installArg}}
}

func executeUpgradeSequence(commands UpgradeSequence) error {
	if env.DryRun {
		log.Info("Running upgrade in dry-run mode...")
	}

	for _, command := range commands {
		err := runCmd(command)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCmd(command UpgradeCommand) error {
	cmd_name := command.Name
	args := command.Args

	_, err := exec.LookPath(cmd_name)

	if err != nil {
		log.Info("Can't find " + cmd_name)
		return err
	}

	cmd := exec.Command(cmd_name, args...)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	log.Infof("Running: %s %s", cmd_name, strings.Join(args, " "))
	if env.DryRun {
		return nil
	} else {
		if err := cmd.Start(); err != nil {
			log.Infof("Was unable to start %s. Error: %v", cmd_name, err)
			return err
		}

		err = cmd.Wait()
		fmt.Println(string(output.Bytes()))

		return err
	}
}
