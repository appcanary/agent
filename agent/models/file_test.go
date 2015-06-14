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

	wfile.StartListener()

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

	cbInvoked := make(chan bool)
	testcb := func(nop *WatchedFile) {
		cbInvoked <- true
	}

	wfile := NewWatchedFile(tf.Name(), testcb)
	wfile.StartListener()
	// File is being wartched
	time.Sleep(200 * time.Millisecond)
	assert.True(wfile.GetBeingWatched())
	os.Remove(tf.Name())
	time.Sleep(200 * time.Millisecond)
	//Since the file is gone, we stopped watching it
	assert.False(wfile.GetBeingWatched())
	wfile.RemoveHook()
}

// does the callback get fired when the directory
// the file is in gets renamed?
func TestWatchFileRenameDirectory(t *testing.T) {
	assert := assert.New(t)

	folder := "/tmp/CANARYTEST"
	file_name := folder + "/test1.gems"

	os.Mkdir(folder, 0777)
	ioutil.WriteFile(file_name, []byte("tst"), 0644)

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
	wfile.StartListener()
	defer wfile.RemoveHook()
	<-cbInvoked

	// aight. let's rename the folder it's in.
	// let's create a tmp path we can rename to.
	folder2 := "/tmp/CANARYTEST2"
	os.Rename(folder, folder2)

	// file should now be missing.
	time.Sleep(500 * time.Millisecond)

	assert.False(wfile.GetBeingWatched())

	// let's then recreate a new file w/same path
	// recreate the old folderm
	os.Mkdir(folder, 0777)
	// write new file
	ioutil.WriteFile(file_name, []byte("tst2"), 0644)

	time.Sleep(500 * time.Millisecond)
	// this file should be different, thus triggering
	// another callback

	mutex.Lock()
	assert.Equal(2, counter)
	mutex.Unlock()

	os.RemoveAll(folder)
	os.RemoveAll(folder2)
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
	wfile.StartListener()
	<-cbInvoked

	// // file gets read on rewrite
	fmt.Println("--> write 2")
	file_content = []byte("hello test - A\n")
	err := ioutil.WriteFile(file_name, file_content, 0644)
	<-cbInvoked

	// we remove and recreate the file,
	// triggering a rehook and re-read
	fmt.Println("--> removal 1")
	os.Remove(file_name)

	fmt.Println("--> write 3")
	file_content = []byte("hello test - AA\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	// we write to the file, triggering
	// another re-read
	fmt.Println("--> write 4")
	file_content = []byte("hello test - AAA\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	// we remove and recreate the file,
	// triggering a rehook yet another re-read
	fmt.Println("--> removal 2")
	os.Remove(file_name)

	fmt.Println("--> write 5")
	file_content = []byte("hello test - AAAA\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-cbInvoked

	fmt.Println("--> write 6")
	file_content = []byte("hello test - AAAAA\n")
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

// TODO: create version of the above test where we compare files that are identical in size, and were touched within one second
