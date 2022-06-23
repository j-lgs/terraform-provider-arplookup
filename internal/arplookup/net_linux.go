package arplookup

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-ping/ping"
	"github.com/mdlayher/arp"
	"inet.af/netaddr"
	"kernel.org/pub/linux/libs/security/libcap/cap"
)

type linuxARP struct {
	dstMAC   net.HardwareAddr
	srcIP    netaddr.IP
	client   *arp.Client
	dropCaps (func() error)
}

func mkLinuxARP(dstMAC net.HardwareAddr) *linuxARP {
	return &linuxARP{
		dstMAC: dstMAC,
	}
}

func isTimeout(err error) bool {
	for {
		if unwrap, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrap.Unwrap()
			continue
		}
		break
	}

	netError, ok := err.(net.Error)

	return ok && netError.Timeout()
}

func (ac *linuxARP) cache(current IP) error {
	if current.cached {
		return nil
	}

	pinger, err := ping.NewPinger(current.String())
	if err != nil {
		return fmt.Errorf("failure creating new `pinger`: %w", err)
	}

	uid := syscall.Getuid()
	if uid == 0 {
		pinger.SetPrivileged(true)
	}

	pinger.Count = 1
	if err := pinger.Run(); err != nil {
		return fmt.Errorf("failure adding IP %s to cache: %w", current.String(), err)
	}

	return nil
}

func (ac *linuxARP) try(chans channels) {
	arp, err := os.Open("/proc/net/arp")
	if err != nil {
		chans.errors <- err
		return
	}
	defer arp.Close()

	// proc arp table has the MAC on field 3 and IP on field 0
	scanner := bufio.NewScanner(arp)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)

		ip := fields[0]
		mac := fields[3]

		if strings.EqualFold(mac, ac.dstMAC.String()) {
			select {
			case <-chans.stop:
				return
			default:
			}

			ip, err := netaddr.ParseIP(ip)
			if err != nil {
				chans.errors <- fmt.Errorf("line: \"%s\" error %w", text, err)
				return
			}

			chans.results <- IP{cached: true, IP: ip}
			return
		}
	}
}

func (ac *linuxARP) request(current netaddr.IP) (IP, error) {
	deadline := time.Now().Add(arpRequestDeadline)
	ac.client.SetReadDeadline(deadline)

	for {
		// Create ARP request addressed to desired MAC and IP
		pkt, err := arp.NewPacket(
			arp.OperationRequest,
			ac.client.HardwareAddr(),
			fromNetaddr(ac.srcIP),
			ac.dstMAC,
			fromNetaddr(current))
		if err != nil {
			return IP{}, err
		}
		if err = ac.client.WriteTo(pkt, ac.dstMAC); err != nil {
			return IP{}, err
		}

		// Read the client's socket for a response, and if we time out, break the loop and return a nil IP
		pkt, _, err = ac.client.Read()
		if isTimeout(err) {
			return IP{}, nil
		}
		if err != nil {
			return IP{}, err
		}

		// If we don't recieve a response from the desired MAC or a reply, continue
		if pkt.Operation != arp.OperationReply || !reflect.DeepEqual(pkt.SenderHardwareAddr, ac.dstMAC) {
			continue
		}

		return IP{cached: false, IP: toNetaddr(pkt.SenderIP)}, nil
	}
}

func (ac *linuxARP) initClient(iface *net.Interface) error {
	client, err := arp.Dial(iface)
	if err != nil {
		return err
	}

	ac.client = client

	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}

	if len(addrs) <= 0 {
		return fmt.Errorf("selected network interface has no assigned addresses")
	}

	ipaddr, err := netaddr.ParseIPPrefix(addrs[0].String())
	if err != nil {
		return err
	}

	ac.srcIP = ipaddr.IP()

	return nil
}

// linuxGetCaps provides the process with the correct capabilities needed to perform raw socket operations.
func linuxGetCaps() (func() error, error) {
	orig := cap.GetProc()
	drop := orig.SetProc // on exit drop capabilities

	caps, err := orig.Dup()
	if err != nil {
		return drop, fmt.Errorf("failed to duplicate process capabilities: %w", err)
	}

	if ok, _ := caps.GetFlag(cap.Permitted, cap.NET_RAW); !ok {
		return drop, fmt.Errorf("insufficient privilege to bind to a raw socket - want %q, have %q", cap.NET_RAW, caps)
	}

	if err := caps.SetFlag(cap.Effective, true, cap.NET_RAW); err != nil {
		return drop, fmt.Errorf("unable to set capability: %v", err)
	}

	if err := caps.SetProc(); err != nil {
		return drop, fmt.Errorf("unable to raise capabilities %q: %v", caps, err)
	}

	return drop, nil
}

func (ac *linuxARP) init(iface *net.Interface) error {
	// Not needed if running as Root
	uid := syscall.Getuid()
	if uid == 0 {
		return ac.initClient(iface)
	}

	drop, err := linuxGetCaps()
	ac.dropCaps = drop
	if err != nil {
		return err
	}

	return ac.initClient(iface)
}

func (ac *linuxARP) destroy() error {
	uid := syscall.Getuid()
	if uid != 0 {
		ac.dropCaps()
	}

	return nil
}
