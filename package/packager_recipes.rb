class UbuntuRecipe < Packager
  self.distro = "ubuntu"
  self.releases = "trusty", "precise", "vivid", "utopic", "wily", "xenial", "yakkety", "zesty"
  self.package_type = "deb"
  CONFIG_FILES = {"config/etc/appcanary/dpkg.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
end

# amazon/2015.03 == el/6 so perhaps
# don't use for now.
# class AmazonRecipe < Packager
#   self.distro = "amazon"
#   self.releases = ["2015.03", "2015.09", "2016.03"]
#   self.package_type ="rpm"
#   CONFIG_FILES = {"config/etc/appcanary/rpm.agent.yml" => "/etc/appcanary/agent.yml.sample",
#                   "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
# end

class CentosRecipe < Packager
  self.distro = "centos"
  self.releases =  ["5", "6"]
  self.package_type = "rpm"
end

# right now we only support centos 7, and 
# i want to ship a conf file w/default turned on
class Centos7Recipe < Packager
  self.distro =  "centos"
  self.releases = ["7"]
  self.package_type = "rpm"

  CONFIG_FILES = {"config/etc/appcanary/rpm.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
end


class RedhatRecipe < Packager
  self.distro = "redhat"
  self.releases =  ["6", "7"]
  self.package_type = "rpm"
end

class DebianRecipe < Packager
  self.distro =  "debian"
  self.releases = ["jessie", "wheezy", "squeeze"]
  self.package_type = "deb"
  CONFIG_FILES = {"config/etc/appcanary/dpkg.agent.yml" => "/etc/appcanary/agent.yml.sample",
                  "config/var/db/appcanary/server.yml" => "/var/db/appcanary/server.yml.sample"}
end

class MintRecipe < Packager
  self.distro =  "linuxmint"
  self.releases = ["rosa", "rafaela", "rebecca", "qiana"]
  self.package_type = "deb"
  self.skip_docker = true
end

class FedoraRecipe < Packager
  self.distro =  "fedora"
  self.releases = ["24", "23"]
  self.package_type = "rpm"
  self.skip_docker = true
end
