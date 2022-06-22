#!/bin/sh
# kill parent processes
cleanup() {
    pkill -P $$
}

# Setup signals to kill child processes on exit.
for sig in INT QUIT HUP TERM; do
  trap "
    cleanup
    trap - $sig EXIT
    kill -s $sig "'"$$"' "$sig"
done
trap cleanup EXIT

if type lsb_release >/dev/null 2>&1 ; then
   DISTRO=$(lsb_release -i -s)
elif [ -e /etc/os-release ] ; then
   DISTRO=$(awk -F= '$1 == "ID" {print $2}' /etc/os-release)
fi

DISTRO=$(printf '%s\n' "$DISTRO" | LC_ALL=C tr '[:upper:]' '[:lower:]') \
      TF_ACC=1 \
      unshare --user --map-root-user --net --mount \
      sh -c 'go test -v -cover ./... -v -timeout 120m'

