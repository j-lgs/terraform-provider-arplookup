package testdriver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os/exec"

	"inet.af/netaddr"
)

// netNS defines a network namespace within the virtual test network.
type netNS struct {
	r     *rand.Rand
	hosts []host
	// Must be in the same subnet as the bridge.
	network net.IPNet
}

// init initialises a Linux network namespace.
func (ns *netNS) init(i int, count int) error {
	// create fd for netns

	nsName := fmt.Sprintf("netns%d", i)

	// create veth pair
	veth := fmt.Sprintf("veth%d", i)
	vethp := fmt.Sprintf("%sp", veth)

	cmds := []*exec.Cmd{
		exec.Command("ip", "netns", "add", nsName),
		exec.Command("ip", "link", "add", veth, "type", "veth", "peer", "name", vethp, "netns", nsName),
		exec.Command("ip", "link", "set", veth, "up"),
		exec.Command("ip", "link", "set", veth, "master", bridge),
	}

	for _, cmd := range cmds {
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running command \"%s\": %w: %s", cmd.String(), err, out.String())
		}
	}

	// create network for ns
	prefix, err := netaddr.ParseIPPrefix(namespaceNetworks[i])
	if err != nil {
		return fmt.Errorf("error parsing CIDR notation network \"%s\": %w", namespaceNetworks[i], err)
	}

	ns.network = *prefix.IPNet()

	if err := ns.initHosts(ns.r, *prefix.IPNet(), count, vethp, nsName); err != nil {
		return err
	}

	return nil
}

// hosts generates count hosts
func (ns *netNS) initHosts(r *rand.Rand, network net.IPNet, count int, peer string, netns string) (err error) {
	ns.hosts = make([]host, count)

	IPNets, err := randIPNets(r, network, count)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		h := host{}
		h.r = r
		h.init(IPNets[i], peer, netns)
		ns.hosts[i] = h
	}

	return nil
}

// randIPNets returns a slice of random /24 addresses inside the `network` subnet
func randIPNets(r *rand.Rand, network net.IPNet, count int) (IPNets []net.IPNet, err error) {
	IPNets = make([]net.IPNet, count)

	if count >= 256 {
		return nil, fmt.Errorf("randIPNets should be called with a count larger than 256")
	}

	perms := r.Perm(253)

	for i := 0; i < len(IPNets); i++ {
		ip := network
		ip.Mask = net.IPv4Mask(0xff, 0xff, 0x80, 0x00) // IPs will be in a /24

		ip.IP[3] = byte(2 + perms[i])

		IPNets[i] = ip
	}

	return
}
