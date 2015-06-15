require 'rake/clean'

def shell(str)
	puts str
	`#{str}`.strip
end

task :default => :build

task :build_all => [:setup, :build]

task :build do
	flags = "-ldflags \"-X main.CanaryVersion #{@release_version || "unreleased" }\""
	`go build #{flags} -o ./bin/canary-agent`
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

	date = `date -u +"%Y.%m.%d-%H%M%S-%Z"`
	tag_name = "deploy-#{date}"
	sha = shell %{git rev-parse --short HEAD}
	user = `whoami`
	commit_message = "#{user} deployed #{sha}"

	@release_version = tag_name

	shell %{git tag -a #{tag_name} -m "#{commit_message}"}
	shell %{git push origin #{tag_name}}
end

task :release => [:release_prep, :default]

task :setup do
	`mkdir -p ./bin`
	`rm -f ./bin/*`
end

