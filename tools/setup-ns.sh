#!/bin/sh

mk_netns () {
    ip netns add "$1"
    ip link add "$2" type veth peer name "$3" netns "$1"
    ip link set "$2" up
    ip addr add "$5" dev "$2"

    ip netns exec "$1" ip link set "$3" up
    ip netns exec "$1" ip addr add "$4" dev "$3"
}

interface="ceth5"
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

# create namespaces and pairs
ip link set lo up
mk_netns netns0  veth0  ceth0  "172.18.0.11/24" "172.18.0.10/31"
mk_netns netns1  veth1  ceth1  "172.18.0.13/24" "172.18.0.12/31"
mk_netns netns2  veth2  ceth2  "172.18.0.15/24" "172.18.0.14/31"
mk_netns netns3  veth3  ceth3  "172.18.0.17/24" "172.18.0.16/31"
mk_netns netns4  veth4  ceth4  "172.18.0.19/24" "172.18.0.18/31"
mk_netns netns5  veth5  ceth5  "172.18.0.21/24" "172.18.0.20/31"
mk_netns netns6  veth6  ceth6  "172.18.0.23/24" "172.18.0.22/31"
mk_netns netns7  veth7  ceth7  "172.18.0.25/24" "172.18.0.24/31"
mk_netns netns8  veth8  ceth8  "172.18.0.27/24" "172.18.0.26/31"
mk_netns netns9  veth9  ceth9  "172.18.0.29/24" "172.18.0.28/31"
mk_netns netns10 veth10 ceth10 "172.18.0.31/24" "172.18.0.30/31"

macaddr=$(ip netns exec "$ns" ip link show "$interface" | grep ether | awk '{ print $2 }')
ip=$(ip netns exec "$ns" ip a show "$interface" | grep inet | head -n1 | awk '{ print $2 }' | cut -f1 -d"/")
TF_ACC=1 MAC="$macaddr" IP="$ip" go test -v -cover ./internal/arplookup -v $1 -timeout 120m
ip netns delete netns0
