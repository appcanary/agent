#!/usr/bin/env bash
set -xeo pipefail

unknown_os() {
  echo -n unknown
}

print_version() {
  local distro=${1} release=${2}
  echo -n ${distro}/${release}
  exit 0
}

detect_linux() {

  # These two contain env variables, but they may not have the ones we need, so
  # we load them in first and then check if we got the variables we wanted
  if [[ -e /etc/os-release ]]; then
    source /etc/os-release
  fi
  if [[ -e /etc/lsb-release ]]; then
    source /etc/lsb-release
  fi

  if [[ (-n "${ID}") && (-n "${VERSION_ID}")]]; then
    distro=${ID}
    release=${VERSION_ID}

    print_version ${distro} ${release}

  elif [[ (-n "${DISTRIB_ID}") && (-n "${DISTRIB_RELEASE}")]]; then
    distro=${DISTRIB_ID}
    release=${DISTRIB_RELEASE}

    print_version ${distro} ${release}

  elif which lsb_release; then
    distro=$(lsb_release -i | cut -f2)
    release=$(lsb_release -r | cut -f2)

    print_version ${distro} ${release}


  elif [[ -e /etc/debian_version ]]; then
    # some Debians have jessie/sid in their /etc/debian_version
    # while others have '6.0.7'
    distro=$(cat /etc/issue | head -1 | awk '{ print tolower($1) }')

    if grep -q '/' /etc/debian_version; then
      release=$(cut --delimiter='/' -f1 /etc/debian_version)
    else
      release=$(cut --delimiter='.' -f1 /etc/debian_version)
    fi

    print_version ${distro} ${release}

  elif [[ -e /etc/oracle-release ]]; then
    release=$(cut -f5 --delimiter=' ' /etc/oracle-release | awk -F '.' '{ print $1 }')
    distro='ol'

    print_version ${distro} ${release}

  elif [[ -e /etc/fedora-release ]]; then
    release=$(cut -f3 --delimiter=' ' /etc/fedora-release)
    distro='fedora'

    print_version ${distro} ${release}

  elif [[ -e /etc/redhat-release ]]; then
    distro_hint=$(cat /etc/redhat-release  | awk '{ print tolower($1) }')
    if [[ "${distro_hint}" = "centos" ]]; then
      release=$(cat /etc/redhat-release | awk '{ print $3 }' | awk -F '.' '{ print $1 }')
      distro='centos'
    elif [[ "${distro_hint}" = "scientific" ]]; then
      release=$(cat /etc/redhat-release | awk '{ print $4 }' | awk -F '.' '{ print $1 }')
      distro='scientific'
    else
      release=$(cat /etc/redhat-release  | awk '{ print tolower($7) }' | cut -f1 --delimiter='.')
      distro='redhatenterpriseserver'
    fi

    print_version ${distro} ${release}

  elif grep -q Amazon /etc/issue; then
    release='6'
    distro='aws'
    print_version ${distro} ${release}
  fi
  unknown_os
}

detect_linux
