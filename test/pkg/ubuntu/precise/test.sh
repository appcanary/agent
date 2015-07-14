#!/usr/bin/env bash
# curl -s https://appcanary.com/servers/install.sh | bash
cd /root
dpkg -i latest.deb
appcanary -version && echo OKAC VERSION
cat /var/log/appcanary.log && echo OKAC LOG
cat /var/db/appcanary/server.conf && echo OKAC SERVERCONF
cat /etc/appcanary/agent.conf && echo OKAC CONFCONF
