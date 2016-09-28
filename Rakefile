require 'rake/clean'
require 'json'
require 'yaml'
load 'test/pkg/Rakefile'

CURRENT_VERSION = "0.1.0"
PC_USER = "appcanary"
PC_REPO = "agent"
PC_STAGING_REPO = "appcanary-stg"

@built_packages = []

def production?
  @isproduction ||= (ENV["CANARY_ENV"] == "production")
end

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

desc "Build the program into ./bin/appcanary"
task :build do
  @ldflags = %{"-X main.CanaryVersion=#{@release_version || "unreleased"}"}
  # actually, do we need to run this every time? let's not for now.
  # shell "go-bindata -pkg detect -o agent/detect/bindata.go agent/resources/"
  shell "go build -ldflags #{@ldflags} -o ./bin/appcanary"
end

desc "Build and run all go tests"
task :test => :build_all do 
	sh "go test -v ./... -race -timeout 20s"
end

desc "Build and run a specific go test"
task :testr => :build_all do
	sh "go test -v ./... -race -timeout 20s -run #{ENV["t"]}"
end

desc "Generate release version from date"
task :release_prep do
  if production?
	  if `git diff --shortstat` != ""
      puts "Whoa there, partner. Dirty trees can't deploy. Git yourself clean"
      exit 1
	  end
  end

	@date = `date -u +"%Y.%m.%d-%H%M%S-%Z"`.strip
	@release_version = "#{CURRENT_VERSION}-#{@date}"
end

desc "Cross compile a binary for every architecture"
task :cross_compile => :release_prep do
  puts "\n\n\n#################################"
  puts "Cross compiling packages."
  puts "#################################\n\n\n"

  @ldflags = %{-X main.CanaryVersion '#{@release_version}'}
  shell %{goxc -build-ldflags="#{@ldflags}" -arch="amd64,386" -bc="linux" -os="linux" -pv="#{@release_version}"  -d="dist/" xc}
end


desc "Generate a package archive for every operating system we support"
task :package => :cross_compile do
  load 'package/packager.rb'
  puts "\n\n\n#################################"
  puts "Building packages."
  puts "#################################\n\n\n"

  [UbuntuRecipe, CentosRecipe, Centos7Recipe, DebianRecipe, MintRecipe, FedoraRecipe].each do |rcp|
    puts "#######"
    puts "#{rcp.distro}"
    puts "#######\n\n"

    pkger = rcp.new(@release_version).build_packager
    @built_packages = @built_packages + pkger.build_packages
  end
end

desc "Cross compile, package and deploy packages to package cloud"
task :deploy => "integration:test" do

  publisher = nil
  if production?
    publisher = PackagePublisher.new(PC_USER, PC_REPO)
  else
    publisher = PackagePublisher.new(PC_USER, PC_STAGING_REPO)
  end

  @built_packages.each do |pkg|
    publisher.publish!(pkg)
  end

  sha = shell %{git rev-parse --short HEAD}
  user = `whoami`.strip
  commit_message = "#{user} deployed #{sha}"


  if production?
    shell %{git tag -a #{@release_version} -m "#{commit_message}"}
    shell %{git push origin #{@release_version}}
  end

end

task :release => [:release_prep, :default, :package]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end
