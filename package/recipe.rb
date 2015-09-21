# TODO: make this play nice with @dont_publish
PC_USER = "appcanary"
PC_REPO = "agent"

class Recipe
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

    def pc_distro_name(name)
      @pc_distro_name = name
    end

    def build!(version, date)
      recipe = self.new
      recipe.distro_name = @distro_name
      recipe.distro_versions = @distro_versions
      recipe.package_type = @package_type
      recipe.version = "#{version}-#{date}"
      recipe.date = date
      recipe.pc_distro_name = @pc_distro_name
      recipe.build!
      recipe
    end

  end
  CONFIG_FILES = {"config/etc/appcanary/agent.conf" => "/etc/appcanary/agent.conf",  "config/var/db/appcanary/server.conf" => "/var/db/appcanary/server.conf"}
  DIRECTORIES = ["/etc/appcanary/", "/var/db/appcanary/"]
  ARCHS = ["amd64", "i386"]
  LICENSE = "GPLv3"
  VENDOR = "appCanary"
  NAME = "appcanary"

  attr_accessor :distro_name, :distro_versions, :package_type, :version, :path, :date
  attr_accessor :releases

  def filename
    "appcanary_0.0.1_#{@arch}_#{@distro}.#{@package_type}"
  end

  def dir_args
    DIRECTORIES.map { |f| "--directories #{f}"}.join(" ")
  end

  def config_files
    self.class::CONFIG_FILES.map {|k, v| "../../../#{k}=#{v}" }.join(" ")
  end

  def list_config_files
    self.class::CONFIG_FILES.map { |k, v| "--config-files #{v} "}.join(" ")
  end

  def full_distro_name(distro_version)
    "#{distro_name}_#{distro_version}"
  end

  def release_path(arch, distro_version)
    "releases/appcanary_#{version}_#{arch}_#{full_distro_name(distro_version)}.#{package_type}" 
  end

  # huge smell right here, gotta fix this
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

  def after_remove_files(dv)
    "--after-remove ./#{package_dir(dv)}/post-remove.sh"
  end

  def after_upgrade_files(dv)
    "--after-upgrade ./#{package_dir(dv)}/post-upgrade.sh"
  end

  def package_files(distro_version)
    "#{package_dir(distro_version)}/files"
  end

  def package_dir(distro_version)
    "package/#{distro_name}/#{distro_version}"
  end

  # the source of all magic
  def build!
    distro_versions.each do |dv|
      ARCHS.each do |arch|
        exec %{bundle exec fpm -f -s dir -t #{package_type} -n #{NAME} -p #{release_path(arch, dv)} -v #{version} -a #{arch} --rpm-os linux -C #{package_files(dv)}  #{dir_args} #{post_install_files(dv)} --license #{LICENSE} --vendor #{VENDOR} #{list_config_files} ./ #{bin_file(arch)} #{config_files}}
        add_release dv, release_path(arch, dv)
      end
    end
  end

  def pc_distro_name
    @pc_distro_name || distro_name 
  end

  def pc_distro_name=(name)
    @pc_distro_name = name
  end

  def releases
    @releases ||= []
  end

  def add_release(dv, path)
    releases << [dv, path]
  end

  def publish!
    releases.each do |dv, rls|
      exec %{bundle exec package_cloud push #{PC_USER}/#{PC_REPO}/#{pc_distro_name}/#{dv} #{rls}}
    end
  end

end

class UbuntuRecipe < Recipe
  distro_name "ubuntu"
  distro_versions "trusty", "precise", "vivid", "utopic"
  package_type "deb"
  # TODO: make this easier to customize
  CONFIG_FILES = {"config/etc/appcanary/ubuntu.agent.conf" => "/etc/appcanary/agent.conf",  "config/var/db/appcanary/server.conf" => "/var/db/appcanary/server.conf"}
end

# amazon/2015.03 == el/6 so perhaps
# don't use for now.
class AmazonRecipe < Recipe
  distro_name "amazon"
  distro_versions "2015.03"
  package_type "rpm"
  pc_distro_name "el"
end

class CentosRecipe < Recipe
  distro_name "centos"
  distro_versions "6", "7"
  package_type "rpm"
  pc_distro_name "el"
end

class RedhatRecipe < Recipe
  distro_name "redhat"
  distro_versions "6", "7"
  package_type "rpm"
end

class DebianRecipe < Recipe
  distro_name "debian"
  distro_versions "jessie", "wheezy", "squeeze"
  package_type "deb"
end

class MintRecipe < Recipe
  distro_name "linuxmint"
  distro_versions "rosa", "rafaela", "rebecca", "qiana" 
  package_type "deb"
end


