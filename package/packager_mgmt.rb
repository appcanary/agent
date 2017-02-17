class Packager
  CONFIG_FILES = {"config/etc/appcanary/agent.yml" => "/etc/appcanary/agent.yml",  
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml"}
  DIRECTORIES = ["/etc/appcanary/", "/var/db/appcanary/"]
  ARCHS = ["amd64", "i386"]

  class << self
    attr_accessor :distro, :releases, :package_type, :version, :skip_docker
  end

  attr_accessor :distro, :releases, :package_type, :version, :skip_docker
  def initialize(version)
    self.version = version

    self.distro = self.class.distro
    self.releases = self.class.releases
    self.package_type = self.class.package_type
    self.skip_docker = self.class.skip_docker
  end

  def build_packages
    releases.map do |rel|
      ARCHS.map do |arch|
        pre_pkg = PrePackage.new(distro, rel, package_type, arch, version, self.class::CONFIG_FILES, DIRECTORIES, skip_docker)
        PackageBuilder.new(pre_pkg).build!
      end
    end.flatten
  end
end

class PackageBuilder
  VENDOR = "Appcanary"
  NAME = "appcanary"
  LICENSE = "GPLv3"

  attr_accessor :package
  def initialize(package)
    self.package = package
  end

  def execute(str)
    puts str
    system str
    puts "\n"
  end

  def build!
    p = self.package
    build_cmd = [
      "bundle exec fpm -f",           # force
      "-s dir",                       # input type
      "-t #{p.package_type}",         # output type
      "-n #{NAME}",                   # package name
      "-p #{p.release_path}",         # where to output
      "-v #{p.version}", 
      "-a #{p.arch}",
      "--rpm-os linux",               # target OS, ignored if not rpm
      "-C #{p.files_path}",           # use this directory to look for files
      "#{p.dir_args}",                # use the following directories when building
      "#{p.post_install_files}",      # after install, use this file
      "--license #{LICENSE} --vendor #{VENDOR}",
      "#{p.list_config_files}",       # mark these files as being config
      "./ #{p.bin_file}",             # where should the binary be copied to?
      "#{p.config_files_path}"        # where are the config files marked above?
    ].join(" ")

    execute build_cmd
    Package.new(p.distro, p.release, p.version, p.release_path, p.skip_docker)
  end
end

class PackagePublisher
  attr_accessor :user, :repo
  def initialize(user, repo)
    self.user = user
    self.repo = repo
  end
 
  def execute(str)
    puts str
    system str
    puts "\n"
  end

  def publish!(pkg)
    execute %{bundle exec package_cloud push #{user}/#{repo}/#{pc_distro_name(pkg.distro)}/#{pkg.release} #{pkg.path}}
  end

  def pc_distro_name(dname)
    case dname
    when "centos"
      "el"
    when "amazon"
      "el"
    else
      dname
    end
  end
end
