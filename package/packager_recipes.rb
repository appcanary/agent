class UbuntuRecipe < PrePackager
  self.distro = "ubuntu"
  self.releases = "trusty", "precise", "vivid", "utopic", "wily", "xenial"
  self.package_type = "deb"
  CONFIG_FILES = {"config/etc/appcanary/ubuntu.agent.conf" => "/etc/appcanary/agent.conf",
                  "config/var/db/appcanary/server.conf" => "/var/db/appcanary/server.conf"}
end

# amazon/2015.03 == el/6 so perhaps
# don't use for now.
class AmazonRecipe < PrePackager
  self.distro = "amazon"
  self.releases = ["2015.03"]
  self.package_type ="rpm"
end

class CentosRecipe < PrePackager
  self.distro = "centos"
  self.releases =  ["5", "6"]
  self.package_type = "rpm"
end

# right now we only support centos 7, and 
# i want to ship a conf file w/default turned on
class Centos7Recipe < PrePackager
  self.distro =  "centos"
  self.releases = ["7"]
  self.package_type = "rpm"

  CONFIG_FILES = {"config/etc/appcanary/rpm.agent.conf" => "/etc/appcanary/agent.conf",
                  "config/var/db/appcanary/server.conf" => "/var/db/appcanary/server.conf"}
end


class RedhatRecipe < PrePackager
  self.distro = "redhat"
  self.releases =  ["6", "7"]
  self.package_type = "rpm"
end

class DebianRecipe < PrePackager
  self.distro =  "debian"
  self.releases = ["jessie", "wheezy", "squeeze"]
  self.package_type = "deb"
end

class MintRecipe < PrePackager
  self.distro =  "linuxmint"
  self.releases = ["rosa", "rafaela", "rebecca", "qiana"]
  self.package_type = "deb"
  self.skip_docker = true
end
