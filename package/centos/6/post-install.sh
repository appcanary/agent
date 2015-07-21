#!/bin/bash

USER="appcanary"

chkconfig appcanary on
if ! id -u $USER > /dev/null 2>&1; then
  useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" $USER
fi
touch /var/log/appcanary.log
chown ${USER}:${USER} /var/log/appcanary.log
