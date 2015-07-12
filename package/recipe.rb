class Recipe
  CURRENT_VERSION = "0.0.1"
  class << self
    def distro_name(name)
      @distro_name = name
    end

    def distro_version(version)
      @distro_version = version
    end

    def package_type(pkg)
      @package_type = pkg
    end

    def post_install(files)
      @post_install = files
    end

    def post_remove(files)
      @post_remove = files
    end

    def post_upgrade(files)
      @post_upgrade = files
    end

    def build(date, arch, path)
      recipe = self.new
      recipe.distro_name = @distro_name
      recipe.distro_version = @distro_version
      recipe.package_type = @package_type
      recipe.post_install = @post_install
      recipe.post_remove = @post_remove
      recipe.post_upgrade = @post_upgrade
      recipe.version = "#{CURRENT_VERSION}-#{date}"
      recipe.date = date
      recipe.arch = arch
      recipe.path = path
      recipe
    end

  end
  CONFIG_FILES = ["/etc/appcanary/agent.conf", "/var/db/appcanary/server.conf"]
  DIRECTORIES = ["/etc/appcanary/", "/var/db/appcanary/"]
  LICENSE = "GPLv3"
  VENDOR = "appCanary"
  NAME = "appcanary"

  attr_accessor :distro_name, :distro_version, :package_type, :post_install, :post_remove, :post_upgrade, :version, :path, :arch, :date

  def filename
    "appcanary_0.0.1_#{@arch}_#{@distro}.#{@package_type}"
  end

  def dir_args
    DIRECTORIES.map { |f| "--directories #{f}"}.join(" ")
  end

  def config_args
    CONFIG_FILES.map {|f| "--config-files #{f}"}.join(" ")
  end

  def release_path
    "releases/appcanary_0.0.1_#{arch}_#{distro_name}.#{package_type}" 
  end

  def bin_path
    "../dist/0.0.1+b#{@date}/linux_#{arch_dir}/appcanary"
  end

  def arch_dir
    # GOXC uses '386' while fpm uses 'i386'. arch => directory it's in
    {"amd64" => "amd64",
     "i386" => "386"}[arch]
  end

  def post_install_files
    "--after-install ./#{package_dir}/post-install.sh"
  end

  def package_files
    "#{package_dir}/files"
  end

  def package_dir
    "package/#{distro_name}/#{distro_version}"
  end


  def build!
    puts %{bundle exec fpm -f -s dir -t #{package_type} -n #{NAME} -p #{release_path} -v #{version} -a #{arch} -C #{package_files} #{config_args} #{dir_args} #{post_install_files} --license #{LICENSE} --vendor #{VENDOR} ./ #{bin_path}=/usr/sbin/appcanary}
  end

end

class UbuntuRecipe < Recipe
  distro_name "ubuntu"
  package_type "deb"
end

# UbuntuRecipe.build!(version, "i386")
