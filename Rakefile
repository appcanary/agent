require 'rake/clean'
require 'json'
require 'yaml'

CURRENT_VERSION = "0.0.2"

@dont_publish = (ENV["CANARY_ENV"] == "test")

# gets result of shell command
def shell(str)
	puts str
	`#{str}`.strip
end

# execute a shell command and print stderr
def exec(str)
  puts str
  system str
end

task :default => :build

task :build_all => [:setup, :build]

task :build do
  @ldflags = %{"-X main.CanaryVersion #{@release_version || "unreleased"}"}
  shell "go build -ldflags #{@ldflags} -o ./bin/canary-agent"
end

task :test => :build_all do 
	sh "go test -v ./... -race -timeout 20s"
end

task :testr => :build_all do
	sh "go test -v ./... -race -timeout 20s -run #{ENV["t"]}"
end

task :release_prep do
  unless @dont_publish
	  if `git diff --shortstat` != ""
      puts "Whoa there, partner. Dirty trees can't deploy. Git yourself clean"
      exit 1
	  end
  end

	@date = `date -u +"%Y.%m.%d-%H%M%S-%Z"`.strip
	tag_name = "#{CURRENT_VERSION}-#{@date}"
	sha = shell %{git rev-parse --short HEAD}
	user = `whoami`.strip
	commit_message = "#{user} deployed #{sha}"

	@release_version = tag_name

  unless @dont_publish
	  shell %{git tag -a #{tag_name} -m "#{commit_message}"}
	  shell %{git push origin #{tag_name}}
  end
end

task :cross_compile => :release_prep do
  @ldflags = %{-X main.CanaryVersion '#{@release_version}'}
  shell %{goxc -build-ldflags="#{@ldflags}" -arch="amd64,386" -bc="linux" -os="linux" -bu="#{@date}"  -d="dist/" xc}
end


task :package => :cross_compile do
  load 'package/recipe.rb'
  [UbuntuRecipe, CentosRecipe, DebianRecipe].each do |rcp|
    
    build = rcp.build!(CURRENT_VERSION, @date)
    unless @dont_publish
      build.publish!
    end
  end
end

task :release => [:release_prep, :default, :package]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end
