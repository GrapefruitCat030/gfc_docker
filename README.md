# gfc_docker
从零学手搓docker

## Linux Namespace

为什么clone出来的sh中,mount | wc 一样:[使用golang理解Linux namespace（五）-Mount](https://here2say.com/post/2019/4/28/go-and-namespace-part5-mount)

强烈建议大家使用 VM + alpine Linux 作为宿主机的方式来做，避免在涉及文件系统挂载的操作时造成灾难性的后果。
