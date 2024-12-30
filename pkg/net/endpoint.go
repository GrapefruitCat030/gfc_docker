package net

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Endpoint struct {
	ID          string
	Device      *netlink.Veth
	MacAddr     net.HardwareAddr
	IPAddr      net.IP
	PortMapping []string
	Network     *Network
}

func (ep *Endpoint) initEndpoint() error {
	ep.Device = &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: ep.ID[:5],
		},
		PeerName: "cif-" + ep.ID[:5],
	}
	if err := netlink.LinkAdd(ep.Device); err != nil { // Only create veth pair, not set up
		return fmt.Errorf("failed to create veth pair: %v", err)
	}
	ep.MacAddr = ep.Device.Attrs().HardwareAddr
	return nil
}

func (ep *Endpoint) configRoute(cinfo *gfc_runinfo.ContainerInfo) error {
	veth_peer, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("failed to lookup veth peer: %v", err)
	}
	// 1. set peer into container ns
	container_pid, err := strconv.Atoi(cinfo.Pid)
	if err != nil {
		return fmt.Errorf("invalid PID: %v", err)
	}
	if err := netlink.LinkSetNsPid(veth_peer, container_pid); err != nil {
		return fmt.Errorf("failed to move veth peer to ns: %v", err)
	}
	// 2. enter container ns
	// Lock the OS thread to prevent the OS from scheduling this goroutine on another thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	container_ns, err := netns.GetFromPid(container_pid)
	if err != nil {
		return fmt.Errorf("failed to get container ns: %v", err)
	}
	defer container_ns.Close()
	original_ns, err := netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get original ns: %v", err)
	}
	defer original_ns.Close()
	if err := netns.Set(container_ns); err != nil {
		return fmt.Errorf("failed to switch to container ns: %v", err)
	}
	defer netns.Set(original_ns)
	// 3. set ip addr
	peerAddr := &netlink.Addr{IPNet: &net.IPNet{IP: ep.IPAddr, Mask: ep.Network.IpRange.Mask}}
	if err := netlink.AddrAdd(veth_peer, peerAddr); err != nil {
		return fmt.Errorf("failed to add addr to veth peer: %v", err)
	}
	// 4. set veth peer and lo up
	if err := netlink.LinkSetUp(veth_peer); err != nil {
		return fmt.Errorf("failed to set veth peer up: %v", err)
	}
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("failed to lookup lo: %v", err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("failed to set lo up: %v", err)
	}
	// 5. set default route
	defaultRoute := &netlink.Route{
		LinkIndex: veth_peer.Attrs().Index,
		Gw:        ep.Network.GatewayIP,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.IPv4Mask(0, 0, 0, 0)},
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("failed to add default route: %v", err)
	}
	return nil
}

func (ep *Endpoint) configPortMapping() error {
	for _, pm := range ep.PortMapping {
		parts := strings.Split(pm, ":")
		if len(parts) != 2 {
			fmt.Printf("port mapping format error: %s\n", pm)
			continue
		}
		hostPort := parts[0]
		containerPort := parts[1]
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s", hostPort, ep.IPAddr.String(), containerPort)
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("failed to exec iptables: %v\noutput: %s\n", err, output)
			continue
		}
	}
	return nil
}
