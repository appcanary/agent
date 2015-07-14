<div style="text-align:center">
  <a href="https://appcanary.com"><img src="https://github.com/appcanary/agent/raw/master/appcanary.png" /></a>
</div>

# Hello.

![circle ci](https://circleci.com/gh/appcanary/agent.png?circle-token=e005a24f2a9e1202caede198cb41d3c09e3eccd6)

This repository holds the source for the [appcanary](https://appcanary.com) agent. 

The agent itself is pretty "dumb". Its sole purpose is to monitor files paths, supplied via config file, for changes and, whenever the file has changed, send it over to the appcanary api for us to parse and make decisions on.

Oh, and it also pings us once an hour so we know it's still working.

If you're reading this because you want to audit the code, the magic starts in [`main.go`](https://github.com/appcanary/agent/blob/master/main.go), [`agent/file.go`](https://github.com/appcanary/agent/blob/master/agent/file.go) and [`agent/agent.go`](https://github.com/appcanary/agent/blob/master/agent/agent.go). We think it's pretty straightforward!

## Installation

1. This project depends on a working golang and ruby environment. 
2. Go to your `$GOPATH`. Type: 

  `go get github.com/appcanary/agent`.
3. `cd` into the brand new agent folder, and type: 

  `go get -t -d -v ./…`
4. Go ahead type: 

  ```bash
gem install bundler # if you don't have it
bundle install
```

This gets you the basic up on your machine.

5. In order to package releases, you're going to need `goxc`, so visit [the goxc github page](https://github.com/laher/goxc) and install that.

6. Finally, [fpm](https://github.com/jordansissel/fpm/) requires `rpmbuild` for rpm packages. On OSX at least, that util is a apart of the `rpm` homebrew package, so:

```bash
brew install rpm
```

## Compiling

Once you've done the above, you're all set!

```bash
rake build # to compile
rake test # to test
rake test t=path/to/test # to test an individual file
```

## Contributing

By submitting a pull request directly to us you hereby agree to assign the copyright of your patch to the Canary Computer Corporation. Our lives are made immensely easier by this and we hope you understand.


![hullo](https://github.com/appcanary/agent/raw/master/readme.gif)
