#!/usr/bin/env sh

if [ -f "/usr/lib/nagios/plugins/check_dhcpv6" ]; then
  chmod u+s /usr/lib/nagios/plugins/check_dhcpv6
fi
