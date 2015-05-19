package models

import (
	"fmt"
	"os"
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
	println("Holla")
	f, _ := os.OpenFile("/tmp/wtf", os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("wtf?!?!?!?!?!?!\n")
	defer f.Close()

	assert := assert.New(t)

	file_content := []byte("tst1")
	tf, _ := ioutil.TempFile("", "gems.lock")
	tf.Write([]byte(file_content))
	tf.Close()
	file_name := tf.Name()

	cbInvoked := make(chan bool)
	invokedCount := make(chan int)
	nextWrite := make(chan bool)
	testcb := func(nop *WatchedFile) {
		cbInvoked <- true
	}

	wfile := NewWatchedFile(file_name, testcb)

	go func() {
		f, _ := os.OpenFile("/tmp/wtf", os.O_APPEND|os.O_WRONLY, 0600)
		defer f.Close()

		timer := time.Tick(3000 * time.Millisecond)
		timer2 := time.Tick(50 * time.Millisecond)
		counter := 0
		for {
			f.WriteString("loop\n")
			select {
			case <-cbInvoked:
				counter++
				nextWrite <- true

			case <-timer2:
				f.WriteString("microloop\n")
			case <-timer:
				f.WriteString("LOOP DED YO\n")
				invokedCount <- counter
				return
			default:
				f.WriteString("default\n")
			}

		}

	}()

	f.WriteString("write1\n")
	// file gets read on hook add
	wfile.AddHook()
	<-nextWrite

	f.WriteString("write2\n")
	// file gets read on rewrite
	fmt.Println("--> write 2")
	file_content = []byte("hello\ntest2\n")
	err := ioutil.WriteFile(file_name, file_content, 0644)
	<-nextWrite

	// we remove and recreate the file,
	// triggering a rehook and re-read
	f.WriteString("delete1\n")
	fmt.Println("--> removal 1")
	os.Remove(file_name)

	f.WriteString("write3\n")
	fmt.Println("--> write 3")
	file_content = []byte("hello\nMOAR\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-nextWrite

	// we write to the file, triggering
	// another re-read
	f.WriteString("write4\n")
	fmt.Println("--> write 4")
	file_content = []byte("hello\ntest3\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-nextWrite

	f.WriteString("delete2\n")

	// we remove and recreate the file,
	// triggering a rehook yet another re-read
	fmt.Println("--> removal 2")
	os.Remove(file_name)

	f.WriteString("write5\n")
	fmt.Println("--> write 5")
	file_content = []byte("hello\ntest lol\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-nextWrite

	f.WriteString("write6\n")
	fmt.Println("--> write 6")
	file_content = []byte("hello\ntest lol3\n")
	err = ioutil.WriteFile(file_name, file_content, 0644)
	assert.Nil(err)
	<-nextWrite
	f.WriteString("write 6 done\n")

	fmt.Println("cleaning up\n")
	// we wrote the file five times, plus the init read
	assert.Equal(6, <-invokedCount)

	f.WriteString("you gotta be kidding me\n")
	// cleanup
	wfile.RemoveHook()
	os.Remove(file_name)
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
