package discovery

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/assert"
)

func TestMembership(t *testing.T) {
	m, handler := setupMember(t, nil)
	m, _ = setupMember(t, m)
	m, _ = setupMember(t, m)

	assert.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			0 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	err := m[2].Leave()
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			serf.StatusLeft == m[0].Members()[2].Status &&
			1 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	assert.Equal(t, fmt.Sprintf("%d", 2), <-handler.leaves)
}

func setupMember(t *testing.T, members []*Membership) (
	[]*Membership, *handler,
) {
	id := len(members)
	ports := Get(1)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
	tags := map[string]string{
		"rpc_addr": addr,
	}

	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}

	h := &handler{}
	if len(members) == 0 {
		h.joins = make(chan map[string]string, 3)
		h.leaves = make(chan string, 3)
	} else {
		c.StartJoinAddrs = []string{
			members[0].BindAddr,
		}
	}

	m, err := New(h, c)
	assert.NoError(t, err)
	members = append(members, m)
	return members, h
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(id, addr string) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}
	return nil
}

func (h *handler) Leave(id string) error {
	if h.leaves != nil {
		h.leaves <- id
	}
	return nil
}

//TODO: Do I need all of this?
//Helper funcs + consts + vars
const (
	lowPort   = 10000
	maxPorts  = 65535
	blockSize = 1024
	maxBlocks = 16
	attempts  = 10
)

var (
	port      int
	firstPort int
	once      sync.Once
	mu        sync.Mutex
	lnLock    sync.Mutex
	lockLn    net.Listener
)

// Get returns n ports that are free to use, panicing if it can't succeed.
func Get(n int) []int {
	ports, err := GetWithErr(n)
	if err != nil {
		panic(err)
	}
	return ports
}

// GetS returns n ports as strings that are free to use, panicing if it can't succeed.
func GetS(n int) []string {
	ports, err := GetSWithErr(n)
	if err != nil {
		panic(err)
	}
	return ports
}

// GetS return n ports (as strings) that are free to use.
func GetSWithErr(n int) ([]string, error) {
	ports, err := GetWithErr(n)
	if err != nil {
		return nil, err
	}
	var portsStr []string
	for _, port := range ports {
		portsStr = append(portsStr, strconv.Itoa(port))
	}
	return portsStr, nil
}

// Get returns n ports that are free to use.
func GetWithErr(n int) (ports []int, err error) {
	mu.Lock()
	defer mu.Unlock()

	if n > blockSize-1 {
		return nil, fmt.Errorf("dynaport: block size is too small for ports requested")
	}

	once.Do(initialize)

	for len(ports) < n {
		port++

		if port < firstPort+1 || port >= firstPort+blockSize {
			port = firstPort + 1
		}

		ln, err := listen(port)
		if err != nil {
			continue
		}
		ln.Close()

		ports = append(ports, port)
	}

	return ports, nil
}

func initialize() {
	if lowPort+maxBlocks*blockSize > maxPorts {
		panic("dynaport: block size is too big or too many blocks requested")
	}
	rand.Seed(time.Now().UnixNano())
	var err error
	for i := 0; i < attempts; i++ {
		block := int(rand.Int31n(int32(maxBlocks)))
		firstPort = lowPort + block*blockSize
		lockLn, err = listen(firstPort)
		if err != nil {
			continue
		}
		return
	}
	panic("dynaport: can't allocated port block")
}

func listen(port int) (*net.TCPListener, error) {
	return net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
}
