package testdriver

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

type Driver struct {
	r  *rand.Rand
	ok bool
	// unique network namespaces for hosts that comprise the test data
	namespaces []*netNS
	bridge     net.Interface
}

func mkBridge() (*net.Interface, error) {
	cmds := []*exec.Cmd{
		exec.Command("ip", "link", "set", "lo", "up"),

		// add bridge and assign an address
		exec.Command("ip", "link", "add", "name", bridge, "type", "bridge"),
		exec.Command("ip", "addr", "add", bridgeAddr, "dev", bridge),
		exec.Command("ip", "link", "set", "dev", bridge, "up"),

		// add tap interface for bridge
		exec.Command("ip", "tuntap", "add", "dev", tap, "mode", "tap"),
		exec.Command("ip", "link", "set", tap, "up"),
		exec.Command("ip", "link", "set", tap, "master", bridge),
	}

	if err := runCmds(cmds); err != nil {
		return nil, err
	}

	return net.InterfaceByName("br0")
}

func remountRun() (err error) {
	distro, ok := os.LookupEnv("DISTRO")
	if !ok {
		return fmt.Errorf("DISTRO not set")
	}
	if distro == "nixos" {
		var system string
		if system, err = os.Readlink("/run/current-system"); err != nil {
			return fmt.Errorf("unable to readlink /run/current-system, got \"%s\": %w", system, err)
		}

		if err = syscall.Mount("none", "/run", "tmpfs", 0, ""); err != nil {
			return fmt.Errorf("unable to bind mount /run to tmpfs: %w", err)
		}

		if err = os.Symlink(system, "/run/current-system"); err != nil {
			return fmt.Errorf("unable to symlink \"%s\" to /run/current-system: %w", system, err)
		}
	} else {
		if err = syscall.Mount("none", "/run", "tmpfs", 0, ""); err != nil {
			return fmt.Errorf("unable to bind mount /run to tmpfs: %w", err)
		}
	}

	os.Mkdir("/run/netns", os.ModeDir)

	return nil
}

// init initialises a testDriver, creating the network environment for this acceptance test to use
func (driver *Driver) Init(r *rand.Rand) error {
	if driver.ok {
		return nil
	}

	uid := syscall.Getuid()
	if uid != 0 {
		return fmt.Errorf("acceptance testing must be performed inside fakeroot user namespace")
	}

	// mount temporary /run
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := remountRun(); err != nil {
		return err
	}

	// ensure namespace is clean
	driver.namespaces = make([]*netNS, nsCount)
	driver.r = r

	bridgeDev, err := mkBridge()
	if err != nil {
		return fmt.Errorf("error getting interface br0 by name: %w", err)
	}
	driver.bridge = *bridgeDev

	for i := 0; i < nsCount; i++ {
		ns := &netNS{r: driver.r}
		if err := ns.init(i, hostCount); err != nil {
			return err
		}
		driver.namespaces[i] = ns
	}

	driver.ok = true

	return nil
}

func randmac(r *rand.Rand) (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}

	// Set local bit
	buf[0] |= 2
	return net.HardwareAddr(buf), nil
}

func (driver *Driver) EnsureNo(mac string) error {
	for _, ns := range driver.namespaces {
		hostlist := []host{}
		for i, host := range ns.hosts {
			if !strings.EqualFold(host.mac.String(), mac) {
				continue
			}

			ns.hosts[i] = ns.hosts[len(ns.hosts)-1]
			hostlist = ns.hosts[:len(ns.hosts)-1]

			mac, err := randmac(driver.r)
			if err != nil {
				return err
			}

			netns := fmt.Sprintf("netns%d", i)
			peer := fmt.Sprintf("veth%dp", i)
			cmds := []*exec.Cmd{
				exec.Command("ip", "netns", "exec", netns, "ifconfig", peer, "hw", "ether", mac.String()),
				exec.Command("ip", "netns", "exec", netns, "ip", "addr", "flush", "dev", peer),
			}

			if err := runCmds(cmds); err != nil {
				return err
			}
		}
		ns.hosts = hostlist
	}

	return nil
}

func (driver *Driver) Needle(mac string, ip string, network string, nsNumber int) error {
	netns := fmt.Sprintf("netns%d", nsNumber)
	peer := fmt.Sprintf("veth%dp", nsNumber)

	cmds := []*exec.Cmd{
		exec.Command("ip", "netns", "exec", netns, "ifconfig", peer, "hw", "ether", mac),
		exec.Command("ip", "netns", "exec", netns, "ip", "addr", "flush", "dev", peer),
		exec.Command("ip", "netns", "exec", netns, "ip", "addr", "add", network, "dev", peer),
	}

	if err := runCmds(cmds); err != nil {
		return err
	}

	return nil
}

func (driver *Driver) NeedleAfter(mac string, ip string, network string, nsNumber int, duration time.Duration) error {
	errs := make(chan error, 1)
	go func(dur time.Duration, errs chan<- error) {
		t := time.NewTimer(dur)
		<-t.C
		if err := driver.Needle(mac, ip, network, nsNumber); err != nil {
			errs <- err
			return
		}
		errs <- nil
	}(duration, errs)

	err := <-errs
	return err
}
