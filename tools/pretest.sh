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

unshare --user --map-root-user --net --mount tools/setup-ns.sh
