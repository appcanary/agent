#!/bin/bash
update-rc.d appcanary defaults
id appcanary > /dev/null 2>&1
if [ $? == 1 ]; then
  useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" appcanary
fi
touch /var/log/appcanary.log
chown appcanary:appcanary /var/log/appcanary.log
