package agent

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockFile struct {
	mock.Mock
}

func (m *MockFile) GetPath() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *MockFile) Parse() interface{} {
	args := m.Mock.Called()
	return args.Get(0)
}

func TestWatchFile(t *testing.T) {
	tf, _ := ioutil.TempFile("", "gemfile")
	filename := tf.Name()
	tf.Write([]byte("lol"))

	app := &App{Name: "test", Path: filename}
	f := new(MockFile)
	f.On("GetPath").Return(tf.Name())
	f.On("Parse").Return("a")
	app.WatchFile(f)
	defer app.CloseWatches()

	//Modify the file to cause a refresh
	tf, _ = os.Open(filename)
	tf.Write([]byte("lol"))
	tf.Close()
	// Sleep to let the watcher catch the refresh
	time.Sleep(1 * time.Second)
	//	tf.
	f.Mock.AssertExpectations(t)
}
