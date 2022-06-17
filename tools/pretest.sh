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

mkdir -p test/bin

slirp_version="1.2.0"
slirp_arch="x86_64"
slirp_sha256="11080fdfb2c47b99f2b0c2b72d92cc64400d0eaba11c1ec34f779e17e8844360"

# Get dependencies
if [ ! -f test/bin/slirp4netns-v${slirp_version} ]; then
    echo "acctest -> downloading slirp4netns binary"
    curl -o test/bin/slirp4netns-v${slirp_version} --fail -L \
	 "https://github.com/rootless-containers/slirp4netns/releases/download/v${slirp_version}/slirp4netns-$(uname -m)"

    echo "$slirp_sha256 test/bin/slirp4netns-v${slirp_version}" | sha256sum -c -;

    chmod +x test/bin/slirp4netns-v${slirp_version}
fi

unshare --user --map-root-user --net --mount sh -c 'sleep 360' &
pid="$!"
sleep 0.1
test/bin/slirp4netns-v${slirp_version} --configure --mtu=65520 --disable-host-loopback "$pid" tap0 > /dev/null 2>&1 &
nsenter -U --wd="$(pwd)" -t "$pid" -m -n --preserve tools/setup-ns.sh
