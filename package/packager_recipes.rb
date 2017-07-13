class UbuntuRecipe < Packager
  self.distro = "ubuntu"
  self.releases = {"precise" => :upstart,
                   "trusty" => :upstart,
                   "utopic" => :upstart,
                   "vivid" => :systemd,
                   "wily" => :systemd,
                   "xenial" => :systemd,
                   "yakkety" => :systemd,
                   "zesty" => :systemd}
  self.package_type = "deb"
  CONFIG_FILES = {"config/etc/appcanary/dpkg.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
end

class CentosRecipe < Packager
  self.distro = "centos"
  self.releases = {"5" => :systemv,
                   "6" => :systemv,
                   "7" => :systemd}

  CONFIG_FILES = {"config/etc/appcanary/rpm.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
  self.package_type = "rpm"
end


class DebianRecipe < Packager
  self.distro =  "debian"
  self.releases = {"jessie" => :systemd,
                   "wheezy" => :systemv,
                   "squeeze" => :systemv}
  self.package_type = "deb"
  CONFIG_FILES = {"config/etc/appcanary/dpkg.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
end

class MintRecipe < Packager
  self.distro =  "linuxmint"
  self.releases = {"rosa" => :upstart,
                   "rafaela" => :upstart,
                   "rebecca" => :upstart,
                   "qiana" => :upstart}
  self.package_type = "deb"
  self.skip_docker = true
end

class FedoraRecipe < Packager
  self.distro =  "fedora"
  self.releases = {"23" => :systemd,
                   "24" => :systemd}
  self.package_type = "rpm"
  self.skip_docker = true
end
