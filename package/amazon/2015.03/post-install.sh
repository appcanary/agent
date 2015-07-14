#!/bin/bash
chkconfig appcanary on
id appcanary > /dev/null 2>&1
if [ $? == 1 ]; then
  useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" appcanary
fi
