#!/bin/dumb-init /bin/sh
set -e

# Modified: https://github.com/hashicorp/docker-consul/blob/2c2873f9d619220d1eef0bc46ec78443f55a10b5/0.X/docker-entrypoint.sh

# If the user is trying to run atlantis directly with some arguments, then
# pass them to atlantis.
if [ "${1:0:1}" = '-' ]; then
    set -- atlantis "$@"
fi

# Look for atlantis subcommands.
if atlantis --help "$1" 2>&1 | grep -q "atlantis $1"; then
    # We can't use the return code to check for the existence of a subcommand, so
    # we have to use grep to look for a pattern in the help output.
    set -- atlantis "$@"
fi

# If we are running atlantis, make sure it executes as the proper user.
if [ "$1" = 'atlantis' ]; then
    # If requested, set the capability to bind to privileged ports before
    # we drop to the non-root user. Note that this doesn't work with all
    # storage drivers (it won't work with AUFS).
    if [ ! -z ${ATLANTIS_ALLOW_PRIVILEGED_PORTS+x} ]; then
        setcap "cap_net_bind_service=+ep" /bin/atlantis
    fi

    set -- gosu atlantis "$@"
fi

exec "$@"