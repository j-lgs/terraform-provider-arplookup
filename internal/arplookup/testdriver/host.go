package testdriver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
)

// host defines a host within the virtual test network.
type host struct {
	r   *rand.Rand
	mac net.HardwareAddr
	ip  net.IPNet
}

// init initalises a new host.
func (hs *host) init(network net.IPNet, peer string, netns string) error {
	hs.ip = network

	// Generate random MAC addresses.
	buf := make([]byte, 6)

	_, err := hs.r.Read(buf)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	// Set local bit
	buf[0] |= 2
	hs.mac = net.HardwareAddr(buf)

	cmds := []*exec.Cmd{
		exec.Command("ip", "netns", "exec", netns, "ip", "link", "set", "dev", "lo", "up"),
		exec.Command("ip", "netns", "exec", netns, "ip", "link", "set", peer, "up"),
		exec.Command("ip", "netns", "exec", netns, "ifconfig", peer, "hw", "ether", hs.mac.String()),
		exec.Command("ip", "netns", "exec", netns, "ip", "addr", "add", hs.ip.String(), "dev", peer),
	}

	for _, cmd := range cmds {
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running command \"%s\": %w: %s", cmd.String(), err, out.String())
		}
	}

	return nil
}
