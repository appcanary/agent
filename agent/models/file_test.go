package models

import (
	"testing"
	"time"

	"io/ioutil"

	"github.com/stateio/testify/assert"
)

// create a tempfile, add a hook, see if hook gets called
// when file changes. TODO: test all other fs events.
func TestWatchFile(t *testing.T) {
	assert := assert.New(t)

	file_content := "tst1"
	tf, _ := ioutil.TempFile("", "gems.lock")
	tf.Write([]byte(file_content))
	tf.Close()

	timer := time.Tick(5 * time.Second)
	cbInvoked := make(chan bool)
	testcb := func(nop *WatchedFile) {
		cbInvoked <- true
	}

	wfile := NewWatchedFile(tf.Name(), testcb)

	wfile.AddHook()

	// let's make sure the file got written to
	read_contents, _ := wfile.Contents()
	assert.Equal(file_content, string(read_contents))

	// but really we want to know if the
	// callback was ever invoked
	select {
	case invoked := <-cbInvoked:
		assert.True(invoked)

	case _ = <-timer:
		assert.True(false)
	}

	// solid. on boot it worked. But what
	// if we changed the file contents?

	newContents := []byte("HelloWorld\n")
	err := ioutil.WriteFile(tf.Name(), newContents, 0777)
	assert.Nil(err)

	// let's wait again just in case.
	select {
	case invoked := <-cbInvoked:
		assert.True(invoked)

	case _ = <-timer:
		assert.True(false)
	}

	wfile.RemoveHook()
}

/*

func TestWatchFile(t *testing.T) {
	tf, _ := ioutil.TempFile("", "gemfile")
	filename := tf.Name()
	tf.Write([]byte("tst1"))
	tf.Close()

	client := &mocks.Client{}
	agent := &Agent{client: client}
	app := &App{Name: "test", Path: filename, callback: agent.Submit}
	f := new(mocks.File)
	f.On("GetPath").Return(filename)

	//We expect Parse to be called three, on first load, and after we modify the file, and after we overwrite it
	f.On("Parse").Return("a").Times(3)
	//We also expect to submit results to the server 3 times
	client.On("Submit", "test", "a").Return(nil).Times(3)
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
	client.Mock.AssertExpectations(t)
}
*/

//TODO: test some pathological cases here
