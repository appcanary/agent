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
	tf.Write([]byte("tst1"))
	tf.Close()

	app := &App{Name: "test", Path: filename}
	f := new(MockFile)
	f.On("GetPath").Return(filename)

	//We expect Parse to be called three, on first load, and after we modify the file, and after we overwrite it
	f.On("Parse").Return("a").Times(3)
	app.WatchFile(f)
	defer app.CloseWatches()

	//Modify the file to cause a refresh
	tf, _ = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0777)
	tf.Write([]byte("tst2"))
	tf.Close()
	time.Sleep(100 * time.Millisecond)

	//Move and overwrite the file to cause a refresh
	os.Rename(filename, filename+".bak")
	tf, _ = os.Create(filename)
	tf.Write([]byte("tst3"))
	tf.Close()
	time.Sleep(200 * time.Millisecond)

	f.Mock.AssertExpectations(t)
}

//TODO: test some pathological cases here
