<div style="text-align:center">
  <a href="https://appcanary.com"><img src="https://github.com/appcanary/agent/raw/master/appcanary-hero.png" /></a>
</div>

# Hello.

![circle ci](https://circleci.com/gh/appcanary/agent.png?circle-token=e005a24f2a9e1202caede198cb41d3c09e3eccd6)

This repository holds the source for the [appcanary](https://appcanary.com) agent. **Please note**: To install the agent on your servers, please consult [the instructions](https://appcanary.com/servers/new). If you're interested in how it works, read on!

The agent itself is pretty "dumb". Its sole purpose is to monitor files paths, supplied via config file, for changes and, whenever the file has changed, send it over to the appcanary api for us to parse and make decisions on.

Oh, and it also pings us once an hour so we know it's still working.

If you're reading this because you want to audit the code, the magic starts in [`main.go`](https://github.com/appcanary/agent/blob/master/main.go), [`agent/file.go`](https://github.com/appcanary/agent/blob/master/agent/file.go) and [`agent/agent.go`](https://github.com/appcanary/agent/blob/master/agent/agent.go). We think it's pretty straightforward!

## Setup

1. This project depends on a working golang and ruby environment, as well as docker.
2. First, let's set up go. Go to your `$GOPATH` and type: 

  `go get github.com/appcanary/agent`

3. `cd` into the brand new agent folder, and install all of our go dependencies by typing: 

  `go get -t -d -v ./…`

4. We'll also need [go-bindata](https://github.com/jteeuwen/go-bindata):

  `go get -u github.com/jteeuwen/go-bindata/...`

  We don't run it by default, cos it's annoying, but whenever you touch a binned file don't forget to run:
```bash
go-bindata -pkg detect -o agent/detect/bindata.go agent/resources/
```
  and check in the result.

5. Now we set bundler and install our ruby dependencies. We use ruby to script all of our build and packaging tasks. Go ahead and type: 

  ```bash
gem install bundler # if you don't have it
bundle install
```

  This gets you all the basics up on your machine.

6. In order to cross compile releases, you're going to need `goxc`, so visit [the goxc github page](https://github.com/laher/goxc) and install that (last we used was version 0.16.0).

7. We package releases using [`fpm`](https://github.com/jordansissel/fpm/). This is installed via bundler in step 4, HOWEVER, `fpm` requires `rpmbuild` in order to assemble rpm packages. We last used `rpmbuild` version 5.4.15. On OSX at least, that util is a apart of the `rpm` homebrew package, so:

  ```bash
brew install rpm
```

8. At this stage you're able to build, test, package and deploy packages. But you know what you're missing? A way to test that the packages work on the (at time of writing) 10 different linux versions you support. We ended up using docker for this. We went and got [boot2docker](http://boot2docker.io/) (cli/docker version 1.6.2 is what we used).

  You may have to also fetch VirtualBox. There's instructions, docker is... complicated.

## Compiling

Once you've done the above, you're all set!

```bash
rake build # to compile
rake test  # to test
rake test t=path/to/test # to test an individual file
```

## Testing on non-Debian based systems

Note that tests will only pass completely on a system with `dpkg` installed. If you do not, we include a `Dockerfile` which should help. There are extra steps needed to get this working. 


Now you should be able to build the docker container and run the tests:

```shell
$ cd $AGENT_REPO_ROOT
$ docker build .
[ ... lots of output elided ... ]
Successfully built <container-id>
```

Where `AGENT_REPO_ROOT` is set to the correct location on your local filesystem and where `<container-sha>` is actually a hexadecimal container ID. Next, run the tests:

```shell
$ docker run <container-id> rake test
```

You will need to rebuild the container for any code changes - don't forget! The container building process caches intermediate layers, so subsequent builds will be faster.

## Packaging

```bash
rake package # just to create packages
rake deploy  # packages, then deploys to package cloud

# actually deploy to 'production' package cloud repo
CANARY_ENV=production rake deploy
```

## Testing the packaging
```bash
boot2docker start # copy and paste the export fields it gives you
rake integration:test
```

or, alternatively, if you built a specific package:

```bash
boot2docker start # again, make sure you copy those exports
rake integration:single distro=debian release=jessie package=releases/appcanary_0.0.2-2015.11.10-212042-UTC_amd64_debian_jessie.deb
```

Pro-tip! Don't forget to use the correct package architecture version for the machine you're on (probably, amd64).

## Contributing

By submitting a pull request directly to us you hereby agree to assign the copyright of your patch to the Canary Computer Corporation. Our lives are made immensely easier by this and we hope you understand.


![hullo](https://github.com/appcanary/agent/raw/master/readme.gif)
