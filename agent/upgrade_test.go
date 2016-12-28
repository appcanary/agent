package agent

import (
	"testing"

	"github.com/stateio/testify/assert"
)

func TestBuildDebianUpgrade(t *testing.T) {
	assert := assert.New(t)

	packageList := map[string]string{"foobar": "version"}
	commands := buildDebianUpgrade(packageList)

	assert.Equal(2, len(commands))
	assert.Equal("apt-get", commands[0].Name)
	assert.Equal("apt-get", commands[1].Name)

	upgradeArgs := commands[1].Args
	lastArg := upgradeArgs[len(upgradeArgs)-1]

	assert.Equal("foobar", lastArg)
}

func TestBuildCentOSUpgrade(t *testing.T) {
	assert := assert.New(t)

	packageList := map[string]string{"foobar": "version"}
	commands := buildCentOSUpgrade(packageList)

	assert.Equal(1, len(commands))
	assert.Equal("yum", commands[0].Name)

	upgradeArgs := commands[0].Args
	lastArg := upgradeArgs[len(upgradeArgs)-1]

	assert.Equal("foobar-version", lastArg)
}
