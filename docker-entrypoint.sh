#!/usr/bin/dumb-init /bin/sh
set -e

# Modified: https://github.com/hashicorp/docker-consul/blob/2c2873f9d619220d1eef0bc46ec78443f55a10b5/0.X/docker-entrypoint.sh

# If the user is trying to run atlantis directly with some arguments, then
# pass them to atlantis.
if [ "$(echo "${1}" | cut -c1)" = "-" ]; then
    set -- atlantis "$@"
fi

# If the user is running an atlantis subcommand (ex. server) then we want to prepend
# atlantis as the first arg to exec. To detect if they're running a subcommand
# we take the potential subcommand and run it through atlantis help {subcommand}.
# If the output contains "atlantis subcommand" then we know it's a subcommand
# since the help output contains that string. For anything else (ex. sh)
# it won't contain that string.
# NOTE: We use grep instead of the exit code since help always returns 0.
if atlantis help "$1" 2>&1 | grep -q "atlantis $1"; then
    # We can't use the return code to check for the existence of a subcommand, so
    # we have to use grep to look for a pattern in the help output.
    set -- atlantis "$@"
fi

# If the current uid running does not have a user create one in /etc/passwd
if ! whoami > /dev/null 2>&1; then
  if [ -w /etc/passwd ]; then
    echo "${USER_NAME:-default}:x:$(id -u):0:${USER_NAME:-default} user:/home/atlantis:/sbin/nologin" >> /etc/passwd
  fi
fi

# If we need to install some tools at entrypoint level, we can add shell scripts
# in folder /docker-entrypoint.d/ with extension .sh and this scripts will be executed
# at entrypount level.
if /usr/bin/find "/docker-entrypoint.d/" -mindepth 1 -maxdepth 1 -type f -print -quit 2>/dev/null | read v; then
  echo "/docker-entrypoint.d/ is not empty, will attempt to perform script execition"
  echo "Looking for shell scripts in /docker-entrypoint.d/"
  find "/docker-entrypoint.d/" -follow -type f -print | sort -V | while read -r f; do
    case "$f" in
      *.sh)
        if [ -x "$f" ]; then
          echo "Launching $f";
          "$f"
        else
          # warn on shell scripts without exec bit
          echo "Ignoring $f, not executable";
        fi
        ;;
      *) echo "Ignoring $f";;
    esac
  done
  echo "Configuration complete; ready for start up"
else
  echo "No files found in /docker-entrypoint.d/, skipping"
fi

# If we're running as root and we're trying to execute atlantis then we use
# gosu (or su-exec which we symlinked to gosu) to step down from root and run as the atlantis user.
if [ "$(id -u)" = 0 ] && [ "$1" = 'atlantis' ]; then
  set -- gosu atlantis "$@"
fi

exec "$@"
