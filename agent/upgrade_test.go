package agent

import (
	"testing"

	"github.com/stateio/testify/assert"
)

func TestBuildDebianUpgrade(t *testing.T) {
	assert := assert.New(t)

	package_list := map[string]string{"foobar": "version"}
	commands := buildDebianUpgrade(package_list)

	assert.Equal(2, len(commands))
	assert.Equal("apt-get", commands[0].Name)
	assert.Equal("apt-get", commands[1].Name)

	upgrade_args := commands[1].Args
	last_arg := upgrade_args[len(upgrade_args)-1]

	assert.Equal("foobar", last_arg)
}
