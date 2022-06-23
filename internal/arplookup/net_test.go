package arplookup

import (
	"context"
	"testing"
	"time"

	"inet.af/netaddr"
)

// TestCheckARPRunTimeout checks wether an empty IP, and errNoIP is returned from checkARPRun if an invalid
// set of IP ranges, in relation to the expected output, is passed to checkARPun. This attempts to model what
// happens with a scan for a machine with a desired MAC doesn't exist is ran.
func TestCheckARPRunInvalid(t *testing.T) {
	var builderIncorrect netaddr.IPSetBuilder
	builderIncorrect.AddPrefix(netaddr.MustParseIPPrefix("192.168.34.0/16"))
	ipSetIncorrect, _ := builderIncorrect.IPSet()

	timeout := 50 * time.Millisecond

	testcases := []struct {
		ipset  *netaddr.IPSet
		expect netaddr.IP
	}{
		{
			ipset:  ipSetIncorrect,
			expect: netaddr.MustParseIP("10.0.33.44"),
		},
		{
			ipset:  ipSetIncorrect,
			expect: netaddr.MustParseIP("10.0.33.44"),
		},
	}

	for _, test := range testcases {
		ac := mkDummyARP(test.expect)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		ip, err := checkARPRun(ctx, ac, ctxData{nil, test.ipset, arpFuncBackoff})

		if err != nil && err != errNoIP {
			t.Fatalf("expected errNoIP from checkARPRun, got: %s", err.Error())
		}

		if !ip.IsZero() {
			t.Fatalf("expected empty IP, got: %s", ip.String())
		}
	}
}

// TestCheckARPRunTimeout verifies whether the function respects the timeout set in the passed context.
func TestCheckARPRunTimeout(t *testing.T) {
	var builderIncorrect netaddr.IPSetBuilder
	builderIncorrect.AddPrefix(netaddr.MustParseIPPrefix("192.168.34.0/16"))
	ipSetIncorrect, _ := builderIncorrect.IPSet()

	testcases := []struct {
		ipset   *netaddr.IPSet
		expect  netaddr.IP
		timeout time.Duration
		backoff time.Duration
	}{
		{
			ipset:   ipSetIncorrect,
			expect:  netaddr.MustParseIP("10.0.33.44"),
			timeout: 50 * time.Millisecond,
			backoff: 5 * time.Second,
		},
		{
			ipset:   ipSetIncorrect,
			expect:  netaddr.MustParseIP("10.0.33.44"),
			timeout: 150 * time.Millisecond,
			backoff: 100 * time.Millisecond,
		},
	}

	for _, test := range testcases {
		ac := mkDummyARP(test.expect)

		ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
		defer cancel()

		start := time.Now()
		checkARPRun(ctx, ac, ctxData{nil, test.ipset, arpFuncBackoff})
		elapsed := time.Since(start)
		if elapsed.Round(2*time.Millisecond) != test.timeout.Round(2*time.Millisecond) {
			t.Fatalf("checkARPRun did not respect context timeout, took \"%s\", should have taken \"%s\"",
				elapsed.String(), test.timeout.String())
		}
	}
}

// TestCheckARPRun checks whether checkARPRun will return valid data if it finds a machine with a desired MAC.
func TestCheckARPRun(t *testing.T) {
	var builderCorrect netaddr.IPSetBuilder
	builderCorrect.AddPrefix(netaddr.MustParseIPPrefix("192.168.33.0/16"))
	ipSetCorrect, _ := builderCorrect.IPSet()

	testcases := []struct {
		ipset     *netaddr.IPSet
		expect    netaddr.IP
		expectErr error
	}{
		{
			ipset:     ipSetCorrect,
			expect:    netaddr.MustParseIP("192.168.33.44"),
			expectErr: nil,
		},
	}

	for _, test := range testcases {
		ac := mkDummyARP(test.expect)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ip, err := checkARPRun(ctx, ac, ctxData{nil, test.ipset, arpFuncBackoff})
		if err != nil && err != test.expectErr {
			t.Fatalf("error encountered while running test: %s", err.Error())
		}

		if ip != test.expect {
			t.Fatalf("expected IP: %s, got: %s", test.expect.String(), ip.String())
		}
	}
}
