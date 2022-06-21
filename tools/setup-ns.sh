#!/bin/sh

mk_netns () {
    ip netns add "$1"
    ip link add "$2" type veth peer name "$3" netns "$1"
    ip link set "$2" up
    #ip addr add "$5" dev "$2"

    ip netns exec "$1" ip link set dev lo up
    ip netns exec "$1" ip link set "$3" up
    ip netns exec "$1" ip addr add "$4" dev "$3"

    ip link set "$2" master "$6"
}

interface="veth5p"
ns="netns5"

if type lsb_release >/dev/null 2>&1 ; then
   DISTRO=$(lsb_release -i -s)
elif [ -e /etc/os-release ] ; then
   DISTRO=$(awk -F= '$1 == "ID" {print $2}' /etc/os-release)
fi

DISTRO=$(printf '%s\n' "$DISTRO" | LC_ALL=C tr '[:upper:]' '[:lower:]')

case "$DISTRO" in
    nixos*)
	SYSTEM="$(readlink /run/current-system)"
	PATH="/run/current-system/sw/bin:$PATH" mount -t tmpfs none /run
	ln -s "$SYSTEM" /run/current-system
	;;
    *)      mount -t tmpfs none /run ;;
esac


# Have bridge between tap and veths
# ...
# profit?

# create namespaces and pairs
ip link set lo up
ip tuntap add dev tap0 mode tap
# ip addr add "10.18.0.1/16" dev tap0
ip link set tap0 up

# setup bridge
ip link add name br0 type bridge
ip addr add "10.18.0.1/16" dev br0
ip link set dev br0 up

ip link set tap0 master br0

mk_netns netns0  veth0  veth0p  "10.18.0.11/16"   "10.18.0.2/24" br0
mk_netns netns1  veth1  veth1p  "10.18.1.11/16"	 "10.18.1.1/24"  br0
mk_netns netns2  veth2  veth2p  "10.18.2.11/16"	 "10.18.2.1/24"  br0
mk_netns netns3  veth3  veth3p  "10.18.3.11/16"	 "10.18.3.1/24"  br0
mk_netns netns4  veth4  veth4p  "10.18.4.11/16"	 "10.18.4.1/24"  br0
mk_netns netns5  veth5  veth5p  "10.18.5.11/16"	 "10.18.5.1/24"  br0
mk_netns netns6  veth6  veth6p  "10.18.6.11/16"	 "10.18.6.1/24"  br0
mk_netns netns7  veth7  veth7p  "10.18.7.11/16"	 "10.18.7.1/24"  br0
mk_netns netns8  veth8  veth8p  "10.18.8.11/17"	 "10.18.8.1/24"  br0
mk_netns netns9  veth9  veth9p  "10.18.9.11/16"	 "10.18.9.1/24"  br0
mk_netns netns10 veth10 veth10p "10.18.10.11/16" "10.18.10.1/24" br0

macaddr=$(ip netns exec "$ns" ip link show "$interface" | grep ether | awk '{ print $2 }')
ip=$(ip netns exec "$ns" ip a show "$interface" | grep inet | head -n1 | awk '{ print $2 }' | cut -f1 -d"/")
TF_ACC=1 MAC="$macaddr" IP="$ip" go test -v -cover ./internal/arplookup -v $1 -timeout 120m
ip netns delete netns0
