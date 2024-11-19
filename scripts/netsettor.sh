#!/usr/bin/env bash

echo 'usage:'$0 '[container_pid]'
echo 'example: sudo bash '$0 '6666'

bridge_nic='brg-demo'
bridge_ip='10.10.10.100/24'

veth_host_ip='10.10.10.100'
veth_container_ip='10.10.10.101/24'
veth_host='veth_host'
veth_container='veth_conta'
pid=$1 #获取传递给脚本的第一个参数

#可能之前有创建过，避免冲突，先移除后添加
ip link del $bridge_nic
ip link add name $bridge_nic address 12:34:56:a1:b2:c3 type bridge

# 添加ip到网桥接口
ip addr add $bridge_ip dev $bridge_nic

ip link set dev $bridge_nic up

#添加虚拟网卡对veth peer
ip link add $veth_host type veth peer name $veth_container

ip link set $veth_host up

#把host端的虚拟网卡附着到个master网卡上（附着在网桥接口）
ip link set $veth_host master $bridge_nic

#虚拟网卡veth另一端移动到容器所在的network namespace
ip link set $veth_container netns $pid


# 暴露容器网络命名空间
NETNS=$pid
if [ ! -d /var/run/netns ]; then
    mkdir /var/run/netns
fi
if [ -f /var/run/netns/$NETNS ]; then
    rm -rf /var/run/netns/$NETNS
fi

ln -s /proc/$NETNS/ns/net /var/run/netns/$NETNS
echo "netns: $NETNS"


# 对容器端虚拟网卡配置网络
ip netns exec $NETNS ip addr add $veth_container_ip dev $veth_container

ip netns exec $NETNS ip link set $veth_container up


ip netns exec $NETNS ip route add default via $veth_host_ip dev $veth_container


# 隐藏容器进程网络命名空间（还原默认设置）
rm -rf /var/run/netns/$NETNS
