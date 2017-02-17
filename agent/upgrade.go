package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/appcanary/agent/conf"
)

type UpgradeCommand struct {
	Name string
	Args []string
}

type UpgradeSequence []UpgradeCommand

func buildCentOSUpgrade(packageList map[string]string) UpgradeSequence {
	installCmd := "yum"
	installArg := []string{"update-to", "--assumeyes"}

	for name, _ := range packageList {
		installArg = append(installArg, strings.TrimSuffix(packageList[name], ".rpm"))
	}

	return UpgradeSequence{UpgradeCommand{installCmd, installArg}}
}

func buildDebianUpgrade(packageList map[string]string) UpgradeSequence {
	env := conf.FetchEnv()

	updateCmd := "apt-get"
	updateArg := []string{"update", "-q"}

	// install only new packages, silence confirm prompt, and
	// if new package has a new set of conf files, update them
	// if the existing conf has not changed from default, or
	// leave old conf in place
	installCmd := "apt-get"
	installArg := []string{"install", "--only-upgrade", "--no-install-recommends", "-y", "-q"}

	if !env.FailOnConflict {
		installArg = append(installArg, "-o Dpkg::Options::=\"--force-confdef\"", "-o Dpkg::Options::=\"--force-confold\"")
	}

	for name, _ := range packageList {
		// for now let's just stick to blanket updates
		// to the packages. At a glance, it seems in ubuntu land you only
		// get access to the most recent version anyways.
		// installArg = append(installArg, name+"="+version)
		installArg = append(installArg, name)
	}

	return UpgradeSequence{UpgradeCommand{updateCmd, updateArg}, UpgradeCommand{installCmd, installArg}}
}

func executeUpgradeSequence(commands UpgradeSequence) error {
	env := conf.FetchEnv()
	log := conf.FetchLog()

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
	log := conf.FetchLog()
	env := conf.FetchEnv()

	cmdName := command.Name
	args := command.Args

	_, err := exec.LookPath(cmdName)

	if err != nil {
		log.Info("Can't find " + cmdName)
		return err
	}

	cmd := exec.Command(cmdName, args...)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	log.Infof("Running: %s %s", cmdName, strings.Join(args, " "))
	if env.DryRun {
		return nil
	} else {
		if err := cmd.Start(); err != nil {
			log.Infof("Was unable to start %s. Error: %v", cmdName, err)
			return err
		}

		err = cmd.Wait()
		fmt.Println(string(output.Bytes()))

		return err
	}
}
