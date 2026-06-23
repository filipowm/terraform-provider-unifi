package testing

import (
	"github.com/apparentlymart/go-cidr/cidr"
	mapset "github.com/deckarep/golang-set/v2"
	"math"
	"net"
	"sync"
	"testing"
)

const (
	vlanMin = 2
	vlanMax = 4095
)

var (
	macInit sync.Once
	macPool = mapset.NewSet[string]()
	network = &net.IPNet{
		IP:   net.IPv4(10, 0, 0, 0).To4(),
		Mask: net.IPv4Mask(255, 0, 0, 0),
	}

	vlanLock sync.Mutex
	vlanNext = vlanMin
)

func GetTestVLAN(t *testing.T) (*net.IPNet, int) {
	vlanLock.Lock()
	defer vlanLock.Unlock()

	vlan := vlanNext
	vlanNext++

	subnet, err := cidr.Subnet(network, int(math.Ceil(math.Log2(vlanMax))), vlan)
	if err != nil {
		t.Error(err)
	}

	return subnet, vlan
}

func AllocateTestMac(t *testing.T) (string, func()) {
	MarkAccTest(t)
	macInit.Do(func() {
		// for test MAC addresses, see https://tools.ietf.org/html/rfc7042#section-2.1.
		// The 00:00:5e:00:53:xx documentation block reserves 256 unicast addresses.
		// Store MAC string VALUES (not pointers) so the set dedupes by value and Pop
		// hands a distinct address to each concurrent caller.
		for i := 0; i < 256; i++ {
			mac := net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, byte(i)}
			if ok := macPool.Add(mac.String()); !ok {
				t.Fatal("Failed to add MAC to pool")
			}
		}
	})

	mac, ok := macPool.Pop()
	if mac == "" || !ok {
		t.Fatal("Unable to allocate test MAC")
	}

	unallocate := func() {
		if ok := macPool.Add(mac); !ok {
			t.Fatal("Failed to add MAC to pool")
		}
	}

	return mac, unallocate
}
