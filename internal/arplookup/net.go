package arplookup

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"time"
	"net"
	"os"
	"strings"

	"github.com/go-ping/ping"
)

type arptable = map[string]net.IP

// arpTableGet defines a function type that returns the system's arp cache as a map of MAC addresses to IPs
type arpTableGet = func() (arptable, error)

type needleType struct {
	mac net.HardwareAddr
	ip  net.IP
}

// mockARPTable4 is a mock implementation of arpTableGet for testing purposes, containing IPv4 addresses and MAC-48 hardware addresses.
// It randomly generates `count` entries. Requires a seed for repeatability on failures. `needle` is a predetermined MAC to be added
// which will randomly be placed in the table.,
func mockARPTable4(seed int64, count int, needle *needleType) arpTableGet {
	return func() (table arptable, err error) {
		table = make(arptable, count)

		IPs := make([]net.IP, count)
		MACs := make([]string, count)

		r := rand.New(rand.NewSource(seed))

		// Generate random IPv4 addresses.
		buf := make([]byte, 4)
		for i, _ := range IPs {
			_, err := r.Read(buf)
			if err != nil {
				return arptable{}, nil
			}

			b := make([]byte, 4)
			copy(b, buf)
			IPs[i] = net.IP(b)
		}

		// Generate random IPv6 MAC addresses.
		buf = make([]byte, 6)
		for i, _ := range MACs {
			_, err := r.Read(buf)
			if err != nil {
				return arptable{}, nil
			}
			// Set local bit
			buf[0] |= 2
			MACs[i] = net.HardwareAddr(buf).String()
		}

		// Put random addresses in table, and put the needle in if it's not nil
		if needle != nil {
			table[needle.mac.String()] = needle.ip
		}
		for i := 0; i < count; i++ {
			table[MACs[i]] = IPs[i]
		}

		return
	}
}

// linuxARPTable is an implementation of arpTableGet for Linux systems.
func linuxARPTable() arpTableGet {
	return func() (table arptable, err error) {
		arp, err := os.Open("/proc/net/arp")
		if err != nil {
			return nil, err
		}
		defer arp.Close()

		// proc arp table has the MAC on field 3 and IP on field 0
		scanner := bufio.NewScanner(arp)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			table[fields[3]] = net.ParseIP(fields[0])
		}

		return
	}
}

// arpTableGet defines a function type that pings a network to refresh the system's ARP cache.
type pollNetIPs = func(network net.IPNet) error

// mockPollIPs is a mock implementation of pollNetIPs for testing purposes.
func mockPollIPs() pollNetIPs {
	return func(net.IPNet) error {
		return nil
	}
}

// Default timeout is 30 seconds
var pollIPTimeout = 30 * time.Second

// linuxPollIPs is an implementation of pollNetIPs for Linux systems.
func linuxPollIPs(ctx context.Context) pollNetIPs {
	return func(network net.IPNet) error {
		ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
		defer cancel()

		err := exec.CommandContext(ctx, "nmap", "-sP", network.String()).Run()
		if err != nil {
			return err
		}
		return nil
	}
}

// macNotFoundError is an error
var ipNotFoundError error = fmt.Errorf("error: IP address corresponding to given MAC address not found in system ARP table")

// checkARP is a wrapper for checkARPRun to abstract out OS specific components.
func checkARP(ctx context.Context, MAC net.HardwareAddr, network net.IPNet) (net.IP, error) {
	return checkARPRun(MAC, network, linuxARPTable(), linuxPollIPs(ctx))
}

// checkARPRun searches an ARP table for a given MAC address in a platform agnostic way. It is important
// to check if the returned error is macNotFoundError to determine the difference between a failure in
// operation and a failure to find the mac in the system's table.
func checkARPRun(MAC net.HardwareAddr, network net.IPNet, arpFunc arpTableGet, pollFunc pollNetIPs) (ip net.IP, err error) {
	err = pollFunc(network)
	if err != nil {
		return nil, err
	}

	arps, err := arpFunc()
	if err != nil {
		return nil, err
	}

	// If an entry isn't present in the system's ARP table the host is either down or the address is incorrect.
	ip = nil
	for tableMAC, tableIP := range arps {
		if strings.EqualFold(MAC.String(), tableMAC) {
			ip = tableIP
			continue
		}
	}
	if ip == nil {
		return ip, ipNotFoundError
	}

	return
}
