require 'rake/clean'
require 'json'
require 'yaml'

CURRENT_VERSION = "0.0.1"

@dont_publish = true

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
  @ldflags = %{-X agent.CanaryVersion '#{@release_version}'}
  shell %{goxc -build-ldflags="#{@ldflags}" -arch="amd64,386" -bc="linux" -os="linux" -bu="#{@date}"  -d="dist/" xc}
end

def distro_files(distro)
  case distro
  when /ubuntu/
    "../init/appcanary.upstart=/etc/init/appcanary.conf"
  when "debian/wheezy"
    "../init/default.wheezy=/etc/defaults/appcanary ../init/init.wheezy=/etc/init.d/appcanary"
  when /jessie|stretch/
    "../init/appcanary.service=/etc/systemd/system/appcanary.service"
  end
end

def post_install(distro)
  if distro =~ /wheezy/
    "--after-install ./init/wheezy-postinstall.sh"
  else
    ""
  end
end

task :package => :cross_compile do

  # GOXC uses '386' while fpm uses 'i386'. arch => directory it's in
  architectures = {"amd64" => "amd64",
                   "i386" => "386"}

  distros = {"deb" => ["ubuntu/vivid", "ubuntu/utopic", "ubuntu/trusty", "ubuntu/precise", "debian/wheezy", "debian/jessie", "debian/stretch"]}

  config_files = ["/etc/appcanary/agent.conf", "/var/db/appcanary/server.conf"]
  config_args = config_files.map {|f| "--config-files #{f}"}.join(" ")

  directories = ["/etc/appcanary/", "/var/db/appcanary/"]
  dir_args = directories.map { |f| "--directories #{f}"}.join(" ")
  license = "GPLv3"
  vendor = "appCanary"
  architectures.each_pair do |arch, arch_dir|
    bin_path = "../dist/#{CURRENT_VERSION}+b#{@date}/linux_#{arch_dir}/appcanary"
    distros.each_pair do |package_type, distro_names|
      distro_names.each do |distro|
        release_path = "releases/appcanary_0.0.1_#{arch}_#{distro.tr('/','-')}.#{package_type}"  
        exec %{bundle exec fpm -f -s dir -t #{package_type} -n appcanary -p #{release_path} -v #{@release_version} -a #{arch} -C ./package/  #{config_args} #{dir_args} #{post_install(distro)} --license GPLv3 --vendor appCanary ./ #{bin_path}=/usr/sbin/appcanary #{distro_files(distro)}}
        unless !@dont_publish
          exec %{package_cloud push appcanary/agent/#{distro} #{release_path}}
        end
      end
    end
  end
end

task :release => [:release_prep, :default, :package]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end
