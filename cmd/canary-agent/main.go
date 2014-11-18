package main

import agent "github.com/mveytsman/canary-agent"

func main() {
	//done := make(chan bool)
	agent := agent.NewAgent("canary.conf")
	defer agent.Close()
	//<-done
}
