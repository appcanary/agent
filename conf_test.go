package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConf(t *testing.T) {
	assert := assert.New(t)

	conf := NewConfFromFile("test_files/test.conf")
	assert.Equal("canary.db", conf.Database.Location)
	assert.Equal("My cool Test", conf.Ruby.Projects[0][0])
	assert.Equal("/Users/maxim/tmp/canary-test", conf.Ruby.Projects[0][1])
}
