# gfc_docker
从零学手搓docker


一些问题
为什么clone出来的sh中,mount | wc 一样:[使用golang理解Linux namespace（五）-Mount](https://here2say.com/post/2019/4/28/go-and-namespace-part5-mount)

强烈建议大家使用 VM + alpine Linux 作为宿主机的方式来做，避免在涉及文件系统挂载的操作时造成灾难性的后果。

TODO:
1. 针对不同的Linux发行版，gfc_docker运行前需要进行各方面检查，如 cgroup 版本、相关命令行工具如iptable下载拉取准备、内核版本检查……等等。
2. 错误处理逻辑优化
3. 整体文档整理
4. 兼容性验证
