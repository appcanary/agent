#!/bin/bash
chkconfig appcanary on
useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" appcanary
