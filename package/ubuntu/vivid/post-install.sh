#!/bin/bash
systemctl enable appcanary
useradd -r -d /var/db/appcanary -s /sbin/nologin -c "AppCanary Agent" appcanary
