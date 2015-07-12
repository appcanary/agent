#!/bin/bash
update-rc.d appcanary defaults
useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" appcanary
