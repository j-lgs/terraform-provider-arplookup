package arplookup

import (
	"context"
	"net"
	"testing"
	"time"

	"inet.af/netaddr"
)

// TestCheckARPRunTimeout checks wether an empty IP, and errNoIP is returned from checkARPRun if an invalid
// set of IP ranges, in relation to the expected output, is passed to checkARPun. This attempts to model what
// happens with a scan for a machine with a desired MAC doesn't exist is ran.
func TestCheckARPRunTimeout(t *testing.T) {
	var builderIncorrect netaddr.IPSetBuilder
	builderIncorrect.AddPrefix(netaddr.MustParseIPPrefix("192.168.34.0/16"))
	ipSetIncorrect, _ := builderIncorrect.IPSet()

	testcases := []struct {
		ipset  netaddr.IPSet
		expect netaddr.IP
	}{
		{
			ipset:  *ipSetIncorrect,
			expect: netaddr.MustParseIP("10.0.33.44"),
		},
	}

	for _, test := range testcases {
		ctx, cancel := context.WithCancel(context.Background())
		ac := mkDummyARP(test.expect)

		time.AfterFunc(50*time.Millisecond, cancel)
		ip, err := checkARPRun(ctx, test.ipset, ac, &net.Interface{})

		if err != nil && err != errNoIP {
			t.Fatalf("expected errNoIP from checkARPRun, got: %s", err.Error())
		}

		if !ip.IsZero() {
			t.Fatalf("expected empty IP, got: %s", ip.String())
		}
	}
}

// TestCheckARPRun checks whether checkARPRun will return valid data if it finds a machine with a desired MAC.
func TestCheckARPRun(t *testing.T) {
	var builderCorrect netaddr.IPSetBuilder
	builderCorrect.AddPrefix(netaddr.MustParseIPPrefix("192.168.33.0/16"))
	ipSetCorrect, _ := builderCorrect.IPSet()

	testcases := []struct {
		ipset     netaddr.IPSet
		expect    netaddr.IP
		expectErr error
	}{
		{
			ipset:     *ipSetCorrect,
			expect:    netaddr.MustParseIP("192.168.33.44"),
			expectErr: nil,
		},
	}

	for _, test := range testcases {
		ac := mkDummyARP(test.expect)

		ip, err := checkARPRun(context.Background(), test.ipset, ac, &net.Interface{})
		if err != nil && err != test.expectErr {
			t.Fatalf("error encountered while running test: %s", err.Error())
		}

		if ip != test.expect {
			t.Fatalf("expected IP: %s, got: %s", test.expect.String(), ip.String())
		}
	}
}
