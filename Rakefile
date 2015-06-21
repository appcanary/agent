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
  @ldflags = %{-X main.CanaryVersion '#{@release_version}'}
  shell %{goxc -build-ldflags="#{@ldflags}" -arch="amd64,386" -bc="linux" -os="linux" -bu="#{@date}"  -d="dist/" xc}
end

task :package => :cross_compile do
  require 'fog'
  aws = YAML.load_file(".aws.yml")
  connection = Fog::Storage.new(
    {:provider                 => 'AWS',
     :aws_access_key_id        => aws['access_key'],
     :aws_secret_access_key    => aws['secret_key'],
     :region                   => 'us-west-1'
    })
  directory = connection.directories.get("appcanary")
  
  ["amd64", "i386"].each do |arch|
    arch_dir =(arch == "i386") ? "386" :  arch #goxc uses 386 not i386
    ["deb", "rpm"].each do |package|
      exec %{fpm -s dir -t #{package} -n canary-agent -p "releases/canary-agent_#{@release_version}_#{arch}.#{package}" -v #{@release_version} -a #{arch} -C ./package/  --config-files /etc/canary-agent/canary.conf --config-files /var/db/canary-agent/server.conf --directories /etc/canary-agent/ --directories /var/db/canary-agent/ --license GPLv3 --vendor canary ./ ../dist/#{CURRENT_VERSION}+b#{@date}/linux_#{arch_dir}/canary-agent=/usr/sbin/canary-agent}
      puts "Uploading to s3: https://appcanary.s3.amazonaws.com/dist/canary-agent_#{@release_version}_#{arch}.#{package}"
      file = directory.files.create(
        :key    => "dist/canary-agent_#{@release_version}_#{arch}.#{package}",
        :body   => File.open("releases/canary-agent_#{@release_version}_#{arch}.#{package}"),
        :public => true
      )
      puts "Uploading to s3: https://appcanary.s3.amazonaws.com/dist/canary-agent_latest_#{arch}.#{package}"
      file.copy("appcanary", "dist/canary-agent_latest_#{arch}.#{package}, :public => true")
    end
  end
end

task :release => [:release_prep, :default, :package]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end

task :test2 do
  system %{fpm -s dir -t rpm -n canary-agent -p "releases/canary-agent_0.0.1-2015.06.16-180218-UTC_386.deb" -v 0.0.1-2015.06.16-180218-UTC -a 386 -C ./package/  --config-files /etc/canary-agent/canary.conf --config-files /var/db/canary-agent/server.conf --directories /etc/canary-agent/ --directories /var/db/canary-agent/ --license GPLv3 --vendor canary ./ ../dist/0.0.1+b2015.06.16-180218-UTC/linux_386/canary-agent=/usr/sbin/canary-agent}
end
