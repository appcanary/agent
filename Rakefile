require 'rake/clean'
require 'json'
require 'yaml'

CURRENT_VERSION = "0.0.1"

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
	`go build -ldflags=#{@ldflags} -o ./bin/canary-agent`
end

task :test => :build_all do 
	sh "go test -v ./... -race -timeout 20s"
end

task :testr => :build_all do
	sh "go test -v ./... -race -timeout 20s -run #{ENV["t"]}"
end

task :release_prep do
	if `git diff --shortstat` != ""
   puts "Whoa there, partner. Dirty trees can't deploy. Git yourself clean"
   exit 1
	end

	@date = `date -u +"%Y.%m.%d-%H%M%S-%Z"`.strip
	tag_name = "#{CURRENT_VERSION}-#{@date}"
	sha = shell %{git rev-parse --short HEAD}
	user = `whoami`.strip
	commit_message = "#{user} deployed #{sha}"

	@release_version = tag_name

	shell %{git tag -a #{tag_name} -m "#{commit_message}"}
	shell %{git push origin #{tag_name}}
end

task :cross_compile => :release_prep do
  @ldflags = %{-X agent.CanaryVersion '#{@release_version}'}
  shell %{goxc -build-ldflags="#{@ldflags}" -arch="amd64,386" -bc="linux" -os="linux" -bu="#{@date}"  -d="dist/" xc}
end

task :package => :cross_compile do
 ["amd64", "i386"].each do |arch|
    arch_dir =(arch == "i386") ? "386" :  arch #goxc uses 386 not i386
    ["deb"].each do |package| # TODO: add "rpm" here when we support it
      exec %{fpm -s dir -t #{package} -n appcanary -p "releases/appcanary_#{@release_version}_#{arch}.#{package}" -v #{@release_version} -a #{arch} -C ./package/  --config-files /etc/appcanary/agent.conf --config-files /var/db/appcanary/server.conf --directories /etc/appcanary/ --directories /var/db/appcanary/ --license GPLv3 --vendor canary ./ ../dist/#{CURRENT_VERSION}+b#{@date}/linux_#{arch_dir}/appcanary=/usr/sbin/appcanary}
    end
 end
 ["ubuntu/vivid", "ubuntu/utopic", "ubuntu/trusty", "ubuntu/precise"].each do |version|
   exec %{package_cloud push appcanary/agent/#{version} releases/appcanary_#{@release_version}*.deb}
 end
end

task :release => [:release_prep, :default, :package]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end
