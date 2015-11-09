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
2. First, let's set up go. Go to your `$GOPATH` and type: 

  `go get github.com/appcanary/agent`

3. `cd` into the brand new agent folder, and install all of our go dependencies by typing: 

  `go get -t -d -v ./…`

4. Now we set bundler and install our ruby dependencies. We use ruby to script all of our build and packaging tasks. Go ahead and type: 

  ```bash
gem install bundler # if you don't have it
bundle install
```

This gets you all the basics up on your machine.

5. In order to cross compile releases, you're going to need `goxc`, so visit [the goxc github page](https://github.com/laher/goxc) and install that (last we used was version 0.16.0).

6. We package releases using [`fpm`](https://github.com/jordansissel/fpm/). This is installed via bundler in step 4, HOWEVER, `fpm` requires `rpmbuild` in order to assemble rpm packages. We last used `rpmbuild` version 5.4.15. On OSX at least, that util is a apart of the `rpm` homebrew package, so:

```bash
brew install rpm
```

## Compiling

Once you've done the above, you're all set!

```bash
rake build # to compile
rake test  # to test
rake test t=path/to/test # to test an individual file
```

## Packaging

```bash
rake package # just to create packages
rake deploy  # packages, then deploys to package cloud

# actually deploy to 'production' package cloud repo
CANARY_ENV=production rake deploy
```

## Contributing

By submitting a pull request directly to us you hereby agree to assign the copyright of your patch to the Canary Computer Corporation. Our lives are made immensely easier by this and we hope you understand.


![hullo](https://github.com/appcanary/agent/raw/master/readme.gif)
