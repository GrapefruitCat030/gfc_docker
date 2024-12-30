package net

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
)

type Network struct {
	Name      string
	IpRange   net.IPNet
	GatewayIP net.IP
	Driver    string
}

func (nw *Network) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string `json:"name"`
		IpRange   string `json:"ip_range"`
		GatewayIP string `json:"gateway_ip"`
		Driver    string `json:"driver"`
	}{
		Name:      nw.Name,
		IpRange:   nw.IpRange.String(),
		GatewayIP: nw.GatewayIP.String(),
		Driver:    nw.Driver,
	})
}

func (nw *Network) UnmarshalJSON(b []byte) error {
	data := &struct {
		Name      string `json:"name"`
		IpRange   string `json:"ip_range"`
		GatewayIP string `json:"gateway_ip"`
		Driver    string `json:"driver"`
	}{}
	if err := json.Unmarshal(b, data); err != nil {
		return err
	}
	_, cidr, err := net.ParseCIDR(data.IpRange)
	if err != nil {
		return err
	}
	nw.Name = data.Name
	nw.IpRange = *cidr
	nw.GatewayIP = net.ParseIP(data.GatewayIP)
	nw.Driver = data.Driver
	return nil
}

// dump writes the network configuration to a file
func (nw *Network) dump(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fp := filepath.Join(path, nw.Name)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(nw)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (nw *Network) load(path string) error {
	fp := filepath.Join(path, nw.Name)
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	data := make([]byte, 4096)
	num, err := f.Read(data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data[:num], nw); err != nil {
		return err
	}
	return nil
}
