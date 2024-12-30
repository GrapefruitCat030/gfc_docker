package net

import (
	"net"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	"github.com/vishvananda/netlink"
)

type Endpoint struct {
	ID          string
	Device      netlink.Veth
	MacAddr     net.HardwareAddr
	IPAddr      net.IP
	PortMapping []string
	Network     *Network
}

func configEndpoint(ep *Endpoint, cinfo *gfc_runinfo.ContainerInfo) error {
	// // 1. 创建veth pair设备
	// if err := createVethPair(ep); err != nil {
	// 	return err
	// }
	// // 2. 设置容器侧veth设备的地址和路由
	// if err := configVeth(ep, cinfo); err != nil {
	// 	return err
	// }
	// // 3. 配置容器端端口映射
	// if err := configPortMapping(ep); err != nil {
	// 	return err
	// }
	return nil
}
