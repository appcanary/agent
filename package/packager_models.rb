class Package
  attr_accessor :distro, :release, :version, :path, :skip_docker
  def initialize(distro, release, version, path, skip_docker)
    self.distro = distro
    self.release = release
    self.version = version
    self.path = path
    self.skip_docker = skip_docker
  end
end

class PrePackage
  attr_accessor :distro, :release, :package_type, :arch, :version, :config_files, :directories, :skip_docker
  def initialize(distro, release, package_type, arch, version, config_files, directories, skip_docker)
    self.distro = distro
    self.release = release
    self.package_type = package_type
    self.arch = arch
    self.version = version
    self.config_files = config_files
    self.directories = directories
    self.skip_docker = skip_docker
  end

  def dir_args
    directories.map { |f| "--directories #{f}"}.join(" ")
  end

  def config_files_path
    config_files.map {|k, v| "../../../#{k}=#{v}" }.join(" ")
  end

  def full_distro_name
    "#{distro}_#{release}"
  end

  def release_path
    "releases/appcanary_#{version}_#{arch}_#{full_distro_name}.#{package_type}" 
  end

  # also, remember to document things. why is this
  # four layers deep?
  def bin_path
    "../../../../dist/#{version}/linux_#{arch}/appcanary"
  end

  def bin_file
    "#{bin_path}=/usr/sbin/appcanary"
  end


  def post_install_files
    "--after-install ./#{package_dir}/post-install.sh"
  end

  def after_remove_files
    "--after-remove ./#{package_dir}/post-remove.sh"
  end

  def after_upgrade_files
    "--after-upgrade ./#{package_dir}/post-upgrade.sh"
  end

  def files_path
    "#{package_dir}/files"
  end

  def package_dir
    "package/#{distro}/#{release}"
  end
end
