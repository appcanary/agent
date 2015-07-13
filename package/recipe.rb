class Recipe
  CURRENT_VERSION = "0.0.1"
  class << self
    def distro_name(name)
      @distro_name = name
    end

    def distro_versions(*versions)
      @distro_versions = versions
    end

    def package_type(pkg)
      @package_type = pkg
    end

    def build(date)
      recipe = self.new
      recipe.distro_name = @distro_name
      recipe.distro_versions = @distro_versions
      recipe.package_type = @package_type
      recipe.version = "#{CURRENT_VERSION}-#{date}"
      recipe.date = date
      recipe
    end

  end
  CONFIG_FILES = {"config/etc/appcanary/agent.conf" => "/etc/appcanary/agent.conf",  "config/var/db/appcanary/server.conf" => "/var/db/appcanary/server.conf", "config/var/log/appcanary.log" => "/var/log/appcanary.log"}
  DIRECTORIES = ["/etc/appcanary/", "/var/db/appcanary/"]
  ARCHS = ["amd64", "i386"]
  LICENSE = "GPLv3"
  VENDOR = "appCanary"
  NAME = "appcanary"

  attr_accessor :distro_name, :distro_versions, :package_type, :post_install, :post_remove, :post_upgrade, :version, :path, :date

  def filename
    "appcanary_0.0.1_#{@arch}_#{@distro}.#{@package_type}"
  end

  def dir_args
    DIRECTORIES.map { |f| "--directories #{f}"}.join(" ")
  end

  def config_files
    CONFIG_FILES.map {|k, v| "../../../#{k}=#{v}" }.join(" ")
  end

  def release_path(arch, distro_version)
    "releases/appcanary_0.0.1_#{arch}_#{full_distro_name(distro_version)}.#{package_type}" 
  end

  def bin_path(arch)
    "../../../../dist/0.0.1+b#{@date}/linux_#{arch_dir(arch)}/appcanary"
  end

  def bin_file(arch)
    "#{bin_path(arch)}=/usr/sbin/appcanary"
  end

  def arch_dir(arch)
    # GOXC uses '386' while fpm uses 'i386'. arch => directory it's in
    {"amd64" => "amd64",
     "i386" => "386"}[arch]
  end

  def post_install_files(distro_version)
    "--after-install ./#{package_dir(distro_version)}/post-install.sh"
  end


  def package_files(distro_version)
    "#{package_dir(distro_version)}/files"
  end

  def package_dir(distro_version)
    "package/#{distro_name}/#{distro_version}"
  end

  def full_distro_name(distro_version)
    "#{distro_name}-#{distro_version}"
  end

  def build!
    distro_versions.each do |dv|
      ARCHS.each do |arch|
        puts %{bundle exec fpm -f -s dir -t #{package_type} -n #{NAME} -p #{release_path(arch, dv)} -v #{version} -a #{arch} -C #{package_files(dv)}  #{dir_args} #{post_install_files(dv)} --license #{LICENSE} --vendor #{VENDOR} ./ #{bin_file(arch)} #{config_files}}
      end
    end
  end

end

class UbuntuRecipe < Recipe
  distro_name "ubuntu"
  distro_versions "trusty"
  package_type "deb"
end

# UbuntuRecipe.build!(version, "i386")
