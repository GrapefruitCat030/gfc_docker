package net

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"text/tabwriter"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
)

const (
	defaultNetPath      = "/var/run/gfc_docker/network/"
	defaultNetworkPath  = "/var/run/gfc_docker/network/networks/"
	defaultEndpointPath = "/var/run/gfc_docker/network/endpoints/"
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

func ConnectEndpoint(netname string, cinfo *gfc_runinfo.ContainerInfo) error {
	nw := &Network{Name: netname}
	nw.load(defaultNetworkPath)
	epIP, err := GlobalIPAM().AllocateIP(nw.IpRange)
	if err != nil {
		return err
	}
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, nw.Name),
		IPAddr:      epIP,
		Network:     nw,
		PortMapping: cinfo.PortMapping,
	}
	if err := ep.initEndpoint(); err != nil {
		return err
	}
	if err := globalDrivers[nw.Driver].Connect(nw, ep); err != nil {
		return err
	}
	if err := ep.configRoute(cinfo); err != nil {
		return err
	}
	if err := ep.configPortMapping(); err != nil {
		return err
	}
	return nil
}

func ListNetworks() error {
	networks := make([]*Network, 0)
	err := filepath.Walk(defaultNetworkPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		nw := &Network{Name: info.Name()}
		if err := nw.load(defaultNetworkPath); err != nil {
			return err
		}
		networks = append(networks, nw)
		return nil
	})
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintln(w, "Name\tDriver\tSubnet\tGateway")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", nw.Name, nw.Driver, nw.IpRange.String(), nw.GatewayIP.String())
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func RemoveNetwork(netname string) error {
	nw := &Network{Name: netname}
	if err := nw.load(defaultNetworkPath); err != nil {
		return err
	}
	// 1.IPAM release
	if err := GlobalIPAM().DeleteSubnet(nw.IpRange); err != nil {
		return err
	}
	// 2.Driver release
	if err := globalDrivers[nw.Driver].Delete(nw); err != nil {
		return err
	}
	// 3.Remove network
	return nw.remove(defaultNetworkPath)
}
