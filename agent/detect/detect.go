package detect

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type LinuxOSInfo struct {
	Distro  string
	Release string
}

func loadScript() []byte {
	data, err := Asset("agent/resources/detect_linux.sh")
	if err != nil {
		panic("Can't find detect_linux.sh - something is *wrong*.")
	}
	return data
}

func run(c *exec.Cmd) (string, error) {
	if c.Stdout != nil {
		return "", errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return "", errors.New("exec: Stderr already set")
	}
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	if err != nil {
		return "Detection failed:\n" + string(stderr.Bytes()), err
	} else {
		return string(stdout.Bytes()), err
	}
}

func parseDetectionString(str string) (*LinuxOSInfo, error) {
	if str == "unknown" {
		return nil, errors.New("Unknown linux distro")
	} else {
		str = strings.ToLower(str)
		splat := strings.Split(str, "/")
		return &LinuxOSInfo{Distro: splat[0], Release: splat[1]}, nil
	}
}

func DetectOS() (*LinuxOSInfo, error) {
	script := string(loadScript())

	_, err := exec.LookPath("bash")
	if err != nil {
		return nil, errors.New("Can't find bash.")
	}

	cmd := exec.Command("bash", "-c", script)
	output, err := run(cmd)
	if err != nil {
		return nil, errors.New(output)
	}
	return parseDetectionString(output)

}
