#!/bin/bash -
#
#
unsafessh serv &> /var/log/unsafessh.log &
echo $$ > /var/run/unsafessh.pid
exec "$@"
