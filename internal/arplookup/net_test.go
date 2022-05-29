package arplookup

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func mustParseMAC(t *testing.T, input string) (MAC net.HardwareAddr) {
	MAC, err := net.ParseMAC(input)
	if err != nil {
		t.Fatalf("error parsing MAC address: %s", err.Error())
	}
	return MAC
}

func parseIP(input string) *net.IP {
	ip := net.ParseIP(input)
	return &ip
}

// TestCheckARPRunRand4 tests the checkARPRun function with mock functions. The mock `arptable` function is
// `mockARPTable4 and generates random MAC and IPv4 address data to suppliment unit tests with static datasets.
// Network isn't used here so it will be passed as a null value to the test.
//
// A false value for expect will signal that we shoudn't insert a known value to the mock arptable.
func TestCheckARPRunRand4(t *testing.T) {
	seed := time.Now().UnixNano()

	// Set seed to the value shown in the test to aid in debugging
	// Example: seed = int64(1653732731739890760)

	testcases := []struct {
		mac       net.HardwareAddr
		count     int
		expect    *net.IP
		expectErr error
	}{
		{
			mac:       mustParseMAC(t, "1b:55:91:bc:82:54"),
			count:     20,
			expect:    parseIP("192.168.33.44"),
			expectErr: nil,
		}, {
			mac:       mustParseMAC(t, "1b:55:91:bc:82:54"),
			count:     20,
			expect:    nil,
			expectErr: ipNotFoundError,
		},
	}

	for _, test := range testcases {
		var needle *needleType = nil

		if test.expect != nil {
			needle = &needleType{
				mac: test.mac,
				ip:  *test.expect,
			}
		}

		arpFunc := mockARPTable4(seed, test.count, needle)

		ip, err := checkARPRun(test.mac, net.IPNet{}, arpFunc, mockPollIPs())
		if err != nil && err != test.expectErr {
			t.Fatalf("(seed %d) error encountered while running test: %s", seed, err.Error())
		}
		if test.expect != nil {
			if !reflect.DeepEqual(ip, *test.expect) {
				t.Fatalf("(seed %d) expected IP: %s, got: %s", seed, (*test.expect).String(), ip.String())
			}
		}
	}
}
