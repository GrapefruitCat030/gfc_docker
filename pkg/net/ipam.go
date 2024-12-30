package net

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

func init() {
	if err := loadIPAM(); err != nil {
		fmt.Println("Error loading IPAM:", err)
	}
}

// IPAM represents the IP Address Management
type IPAM struct {
	// 子网 -> IP分配状态的映射
	// key = 子网 string, value = 位图
	Subnets map[string]*bitmap `json:"subnets"`
	// 文件锁，确保并发安全
	sync.RWMutex
}

type bitmap struct {
	// IP分配状态位图
	Bitmap []byte `json:"bitmap"`
	// 子网信息
	Subnet *net.IPNet `json:"subnet"`
}

func (m *bitmap) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Bitmap []byte `json:"bitmap"`
		Subnet string `json:"subnet"`
	}{
		Bitmap: m.Bitmap,
		Subnet: m.Subnet.String(),
	})
}

func (m *bitmap) UnmarshalJSON(b []byte) error {
	data := &struct {
		Bitmap []byte `json:"bitmap"`
		Subnet string `json:"subnet"`
	}{}
	if err := json.Unmarshal(b, data); err != nil {
		return err
	}
	_, cidr, err := net.ParseCIDR(data.Subnet)
	if err != nil {
		return err
	}
	m.Bitmap = data.Bitmap
	m.Subnet = cidr
	return nil
}

var globalIPAM = &IPAM{
	Subnets: make(map[string]*bitmap),
}

func GlobalIPAM() *IPAM {
	return globalIPAM
}

func (i *IPAM) AllocateIP(subnet net.IPNet) (net.IP, error) {
	i.Lock()
	defer i.Unlock()

	key := subnet.String()
	// 做一个深拷贝，避免后续修改
	_, newSubnet, err := net.ParseCIDR(subnet.String())
	if err != nil {
		return nil, err
	}
	// 获取子网对应的位图
	bm, ok := i.Subnets[key]
	if !ok {
		ones, bits := newSubnet.Mask.Size()
		size := 1 << uint(bits-ones)
		bm = &bitmap{
			Bitmap: make([]byte, (size+7)/8), // 向上取整
			Subnet: newSubnet,
		}
		i.Subnets[key] = bm
	}
	// 遍历位图，找到第一个空闲的IP
	// 0是网关IP，最后一个是广播地址
	for i := 1; i < len(bm.Bitmap)*8-1; i++ {
		byteIdx := i / 8
		bitIdx := i % 8
		bitFlag := byte(1 << uint(7-bitIdx))
		if bm.Bitmap[byteIdx]&bitFlag == 0 {
			bm.Bitmap[byteIdx] |= bitFlag
			ip := make(net.IP, len(subnet.IP))
			copy(ip, subnet.IP)
			for j := 0; j < 4; j++ {
				ip[j] |= byte(i >> ((len(ip) - 1 - j) * 8)) // 分段赋值
			}
			// save to file
			if err := saveIPAM(); err != nil {
				return nil, err
			}
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no available IP in subnet %s", key)
}

func (i *IPAM) ReleaseIP(subnet net.IPNet, ipaddr net.IP) error {
	i.Lock()
	defer i.Unlock()
	key := subnet.String()
	bm, ok := i.Subnets[key]
	if !ok {
		return fmt.Errorf("unknown subnet %s", key)
	}

	fmt.Printf("subnet: %v\n", subnet)
	fmt.Printf("ipaddr: %v\n", ipaddr)
	fmt.Printf("befor bm: %v\n", bm)

	// 计算IP地址在位图中的索引
	ipInt := ipToInt(ipaddr)
	subnetInt := ipToInt(subnet.IP)
	idx := int(ipInt - subnetInt)

	fmt.Printf("ipInt: %v\n", ipInt)
	fmt.Printf("subnetInt: %v\n", subnetInt)
	fmt.Printf("idx: %v\n", idx)

	if idx < 0 || idx >= len(bm.Bitmap)*8 {
		return fmt.Errorf("IP %s is out of range", ipaddr)
	}
	// 清除位图中对应的位
	byteIdx := idx / 8
	bitIdx := idx % 8
	flag := byte(1 << uint(7-bitIdx))
	bm.Bitmap[byteIdx] &^= flag

	fmt.Printf("after bm: %v\n", bm)

	// save to file
	return saveIPAM()
}

func (i *IPAM) DeleteSubnet(subnet net.IPNet) error {
	i.Lock()
	defer i.Unlock()
	key := subnet.String()
	if _, ok := i.Subnets[key]; !ok {
		return fmt.Errorf("unknown subnet %s", key)
	}
	delete(i.Subnets, key)
	return saveIPAM()
}

func saveIPAM() error {
	dir := filepath.Dir(defaultIPAMFilePath)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	fp, err := os.OpenFile(defaultIPAMFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := json.NewEncoder(fp).Encode(GlobalIPAM()); err != nil {
		return err
	}
	return nil
}

func loadIPAM() error {
	fp, err := os.OpenFile(defaultIPAMFilePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer fp.Close()
	if err := json.NewDecoder(fp).Decode(GlobalIPAM()); err != nil {
		return err
	}
	return nil
}

// IP地址转整数
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
