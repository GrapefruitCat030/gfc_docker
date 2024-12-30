package net

import (
	"fmt"
	"net"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
)

const (
	defaultNetPath = "/var/run/gfc_docker/network/"

	defaultNetworkPath  = "/var/run/gfc_docker/network/networks"
	defaultEndpointDir  = "endpoints"
	defaultIPAMFilePath = "/var/run/gfc_docker/network/ipam.json"
)

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	gatewayIP, err := GlobalIPAM().AllocateIP(*cidr)
	if err != nil {
		return err
	}
	fmt.Println("gatewayIP:", gatewayIP)
	nw, err := globalDrivers[driver].Create(cidr.String(), gatewayIP.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}

func CreateEndpoint(netname string, cinfo *gfc_runinfo.ContainerInfo) error {
	nw := &Network{Name: netname}
	nw.load(defaultNetworkPath)
	epIP, err := GlobalIPAM().AllocateIP(nw.IpRange)
	if err != nil {
		return err
	}
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", nw.Name, cinfo.Id),
		IPAddr:      epIP,
		Network:     nw,
		PortMapping: cinfo.PortMapping,
	}
	if err := globalDrivers[nw.Driver].Connect(nw, ep); err != nil {
		return err
	}
	if err := configEndpoint(ep, cinfo); err != nil {
		return err
	}
	return nil
}

func ListNetworks() {
	for _, nw := range globalDrivers {
		fmt.Println(nw)
	}
}

func RemoveNetwork(name string) error {
	// 1.IPAM release
	// 2.Driver release
	// 3.Remove network
	return nil
}
