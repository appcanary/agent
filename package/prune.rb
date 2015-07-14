#!/usr/bin/env ruby
# usage: ./package/prune user/repo
#
require 'json'
require 'pp'
require 'rest-client'
require 'time'

API_TOKEN = JSON.parse(`cat ~/.packagecloud`)["token"]
USER_REPO = ARGV[0]

@base_url = "https://#{API_TOKEN}:@packagecloud.io/api/v1/repos/#{USER_REPO}"

package_url = "/packages.json"

url = @base_url + package_url

all_pkg = RestClient.get(url);
pkgs =  JSON.parse(all_pkg)

# prune everything
# pkgs.each do |p|
#   distro_version = p["distro_version"]
#   filename = p["filename"]
#   yank_url = "/#{distro_version}/#{filename}"
#   url = base_url + yank_url
#   puts "yanking #{url}"
#  
#   result = RestClient.delete(url)
#   if result == {}
#     puts "successfully yanked #{filename}!"
#   end
# end


def yank_url(p)
  distro_version = p["distro_version"]
  filename = p["filename"]
  yank_url = "/#{distro_version}/#{filename}"
  @base_url + yank_url
end

pkgs.select { |p| p["version"] < "0.0.2" }.map { |p| RestClient.delete yank_url(p) }
