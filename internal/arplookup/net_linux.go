package arplookup

import (
	"fmt"
	"net"
	"net/netip"
	"reflect"
	"syscall"
	"time"

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

func (ac *linuxARP) request(current netaddr.IP) (netaddr.IP, error) {
	deadline := time.UnixMicro(time.Now().UnixMicro()).Add(arpRequestDeadline)
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
			return netaddr.IP{}, err
		}
		if err = ac.client.WriteTo(pkt, ac.dstMAC); err != nil {
			return netaddr.IP{}, err
		}

		// Read the client's socket for a response, and if we time out, break the loop and return a nil IP
		pkt, _, err = ac.client.Read()
		if isTimeout(err) {
			return netaddr.IP{}, nil
		}
		if err != nil {
			return netaddr.IP{}, err
		}

		// If we don't recieve a response from the desired MAC or a reply, continue
		if pkt.Operation != arp.OperationReply || !reflect.DeepEqual(pkt.SenderHardwareAddr, ac.dstMAC) {
			continue
		}

		return toNetaddr(pkt.SenderIP), nil
	}
}

func (ac *linuxARP) initClient() error {
	iface, err := net.InterfaceByName("br0")
	if err != nil {
		return err
	}

	client, err := arp.Dial(iface)
	if err != nil {
		return err
	}

	ac.client = client

	ipaddr, err := netip.ParseAddr("10.18.0.1")
	if err != nil {
		return err
	}

	ac.srcIP = netaddr.IPFrom4(ipaddr.As4())

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

func (ac *linuxARP) init() error {
	// Not needed if running as Root
	uid := syscall.Getuid()
	if uid == 0 {
		return ac.initClient()
	}

	drop, err := linuxGetCaps()
	ac.dropCaps = drop
	if err != nil {
		return err
	}

	return ac.initClient()
}

func (ac *linuxARP) destroy() error {
	uid := syscall.Getuid()
	if uid != 0 {
		ac.dropCaps()
	}

	return nil
}
