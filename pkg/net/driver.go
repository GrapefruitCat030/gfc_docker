package net

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
)

func init() {
	globalDrivers["bridge"] = &BridgeDriver{}
}

var globalDrivers = make(map[string]NetDriver)

type NetDriver interface {
	Name() string
	Create(subnet, gatewayIp, name string) (*Network, error) // Create a network
	Delete(net *Network) error                               // Delete a network
	Connect(net *Network, endpoint *Endpoint) error          // Connect a endpoint to a network
	Disconnect(net *Network, endpoint *Endpoint) error       // Disconnect a endpoint from a network
}

type BridgeDriver struct {
}

func (bd *BridgeDriver) initDriver(nw *Network) error {
	// 1. create a bridge device
	if _, err := net.InterfaceByName(nw.Name); err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	brg := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: nw.Name,
		},
	}
	if err := netlink.LinkAdd(brg); err != nil {
		return fmt.Errorf("failed to create bridge [%s]: %v", nw.Name, err)
	}
	// 2. set bridge device address and route
	brgLink, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return fmt.Errorf("failed to lookup bridge device: %v", err)
	}
	brgAddr := &netlink.Addr{IPNet: &net.IPNet{IP: nw.GatewayIP, Mask: nw.IpRange.Mask}}
	if err := netlink.AddrAdd(brgLink, brgAddr); err != nil {
		return fmt.Errorf("failed to add gateway address: %v", err)
	}
	// 3. set bridge device up
	if err := netlink.LinkSetUp(brgLink); err != nil {
		return fmt.Errorf("failed to set bridge up: %v", err)
	}
	// 4. set iptables
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", nw.IpRange.String(), nw.Name)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to exec iptables: %v, %s", err, output)
	}
	return nil
}

func (bd *BridgeDriver) Name() string {
	return "bridge"
}

func (bd *BridgeDriver) Create(subnet, gatewayIp, name string) (*Network, error) {
	_, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	nw := &Network{
		Name:      name,
		IpRange:   *ipRange,
		GatewayIP: net.ParseIP(gatewayIp),
		Driver:    bd.Name(),
	}
	if err := bd.initDriver(nw); err != nil {
		return nil, err
	}
	return nw, nil
}

func (bd *BridgeDriver) Delete(net *Network) error {
	brg, err := netlink.LinkByName(net.Name)
	if err != nil {
		return err
	}
	if err := netlink.LinkDel(brg); err != nil {
		return err
	}
	return nil
}

func (bd *BridgeDriver) Connect(net *Network, endpoint *Endpoint) error {
	return nil
}

func (bd *BridgeDriver) Disconnect(net *Network, endpoint *Endpoint) error {
	return nil
}
