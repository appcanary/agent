package models

import (
	"fmt"
	"os"
	"sync"
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

func TestWatchFileFailure(t *testing.T) {
	assert := assert.New(t)

	file_content := "tst1"
	tf, _ := ioutil.TempFile("", "gems.lock")
	tf.Write([]byte(file_content))
	tf.Close()

	// timer := time.Tick(5 * time.Second)
	cbInvoked := make(chan bool)
	testcb := func(nop *WatchedFile) {
		cbInvoked <- true
	}

	wfile := NewWatchedFile(tf.Name(), testcb)

	assert.NotPanics(func() {
		wfile.AddHook()
		os.Remove(tf.Name())
		time.Sleep(200 * time.Millisecond)
	})
	assert.True(true)
	wfile.RemoveHook()
}

func TestWatchFileHookLoop(t *testing.T) {

	assert := assert.New(t)

	file_content := []byte("tst1")
	tf, _ := ioutil.TempFile("", "gems.lock")
	tf.Write([]byte(file_content))
	tf.Close()
	file_name := tf.Name()

	cbInvoked := make(chan bool, 10)

	mutex := &sync.Mutex{}
	counter := 0
	testcb := func(wfile *WatchedFile) {
		mutex.Lock()
		counter++
		mutex.Unlock()
		cbInvoked <- true
	}

	wfile := NewWatchedFile(file_name, testcb)

	// file gets read on hook add
	wfile.AddHook()
	<-cbInvoked

	// file gets read on rewrite
	fmt.Println("--> write 2")
	file_content = []byte("hello test2\n")
	err := ioutil.WriteFile(file_name, file_content, 0644)
	<-cbInvoked

	// we remove and recreate the file,
	// triggering a rehook and re-read
	fmt.Println("--> removal 1")
	os.Remove(file_name)

	fmt.Println("--> write 3")
	file_content = []byte("hello test3\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	// we write to the file, triggering
	// another re-read
	fmt.Println("--> write 4")
	file_content = []byte("hello test4\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	// we remove and recreate the file,
	// triggering a rehook yet another re-read
	fmt.Println("--> removal 2")
	os.Remove(file_name)

	fmt.Println("--> write 5")
	file_content = []byte("hello test5\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	fmt.Println("--> write 6")
	file_content = []byte("hello test6\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	fmt.Println("cleaning up\n")
	// we wrote the file five times, plus the init read
	mutex.Lock()
	assert.True(counter >= 6)
	mutex.Unlock()

	// cleanup
	wfile.RemoveHook()
	os.Remove(file_name)
}
