#!/usr/bin/env bash
set -xeo pipefail

BASEPATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"

LSB_RELEASE="test\/data\/lsb-release"
OS_RELEASE="test\/data\/os-release"
RH_RELEASE="test\/data\/redhat-release"

DETECT_PATH="agent/resources/detect_linux.sh"

test_detect() {
  local replace=${1} with=${2} assert=${3} run_check=""

  run_check=$(cat ${DETECT_PATH} | sed -e "s/${replace}/${with}/" | bash)
  if [[ ${run_check} != ${assert} ]]; then
    echo "FAIL on ${assert} was ${run_check}"
    exit 1
  else
    echo "OK ${assert}"
  fi
}

main() {
  test_detect "\/etc\/lsb-release" ${LSB_RELEASE} "Ubuntu/14.04"
  test_detect "\/etc\/os-release" ${OS_RELEASE} "centos/7"
  test_detect "\/etc\/redhat-release" ${RH_RELEASE} "centos/6"
}

main
