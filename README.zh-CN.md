# Dockin CNI - Dockin Container Network Interface

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

[English](readme.md) | 中文

**更多Dockin组件请访问 [https://github.com/WeBankFinTech/Dockin](https://github.com/WeBankFinTech/Dockin)**

## **Dockin cni**
dockin cni 与资源管理模块（RM）进行交互，共同来管理容器网络，当前支持：
- 创建单网卡网络
- 创建多网卡网络
- 仅支持dockin ipam插件
- 仅支持桥接模式

以下为运行dockin cni的需要组件
- dockin-cni，主要的插件入口，通过调用网桥来进行网络管理，并且与RM模块进行交互
- dockin-ipam，用于分配ip地址
- bridge, 网桥用于网络管理

cni的配置
--
配置示例
```
{
    "cniVersion": "0.2.0",
    "name": "dockin-cni",
    "type": "dockin-cni",
    "confDir": "/etc/cni/dockin/net.d",
    "binDir": "/opt/cni/bin",
    "logFile": "/data/kubernetes/dockin-cni.log",
    "logLevel": "debug",
    "backend": "http://localhost:8080/rm"
}
```
参数描述如下：
- cniVersion, 支持的cni版本
- name, 网络插件的名字
- type, 类型，可执行二进制文件，这里必须使用'dockin-cni'
- confDir, 网络配置文件所在的目录
- binDir, 网桥（bridge）二进制文件所在目录
- logFile, 日志文件路径
- logLevel, 日志登记
- backend, RM模块访问地址

rm 数据示例
---
```
{
    "code": 0,
    "reqId": "1234",
    "message": "success",
    "data": [
        {
            "type": "test",
            "podIp": "192.168.1.2",
            "subnetMask": "255.255.255.0",
            "gateway": "192.168.1.1",
            "ifName": "eth0",
            "master": true
        },
        {
            "type": "dockin",
            "podIp": "192.168.2.2",
            "subnetMask": "255.255.255.0",
            "gateway": "192.168.1.1",
            "ifName": "net0",
            "master": false
        }
    ]
}
```
其中:
- code, 返回码，0表示成功，其余表示失败
- message, 返回的描述信息，包括成功信息和失败信息
- data, 关于网络信息的数据内容
    - type, 网络类型 
    - podIp, 为pod分配的ip地址
    - subnetMask, 子网掩码
    - gateway, 网管
    - ifName, 该网络所属的网卡名称，将能够通过ifconfig命令及ip a命令查看
    - master, 用于标记是否为主要网络，在使用kubectl展示信息时将会看到该网络信息，并且在一个pod中只能有一个master网卡

network 配置
---
network 配置通过网桥进行管理，更多细节可参考以下链接：
>https://github.com/containernetworking/plugins/tree/master/plugins/main/bridge

网络配置通过json文件进行存储，存放在binDir中（在cni配置中），并且将会传给kubelet创建网络。

```
{
  "cniVersion": "0.2.0",
  "name": "dockin",
  "type": "bridge",
  "bridge": "br1"
}
``` 

- cniVersion, 该cni支持的版本
- name, 网络名称，比如与rm中的名称保持一致
- type, 类型，仅支持通过网桥进行网络管理
- bridge, 网桥名称，多网卡环境下可以分配不同的网桥名称

---
## **Dockin-ipam**: 静态IP地址管理插件

### 简介

静态IPAM插件用来为容器分配静态IP地址（IPv4/IPv6），使用与需要为容器分配静态IP的场景（即重启或kill之后ip地址保持不变）

static IPAM is very simple IPAM plugin that assigns IPv4 and IPv6 addresses statically to container. This will be useful in debugging purpose and in case of assign same IP address in different vlan/vxlan to containers.


### 配置示例

```
{
	"ipam": {
		"type": "static",
		"addresses": [
			{
				"address": "192.168.0.1/24",
				"gateway": "192.168.0.254"
			},
			{
				"address": "3ffe:ffff:0:01ff::1/64",
				"gateway": "3ffe:ffff:0::1"
			}
		],
		"routes": [
			{ "dst": "0.0.0.0/0" },
			{ "dst": "192.168.0.0/16", "gw": "192.168.1.1" },
			{ "dst": "3ffe:ffff:0:01ff::1/64" }
		],
		"dns": {
			"nameservers" : ["8.8.8.8"],
			"domain": "example.com",
			"search": [ "example.com" ]
		}
	}
}
```

### 网络配置描述

* `type` (string, 必须): "static"
* `addresses` (array, 可选): ip地址对象数组:
	* `address` (string, 必须): 以CIDR表示的ip地址.
	* `gateway` (string, 可选): 执行子网内的ip作为网管.
* `routes` (string, 可选): 需要加载到容器命名空间的路由列表。每条路由信息都是有dst及gw（可选）字段构成的字典，如果设置了gw字段，“gateway”标签对应的值将会被使用.
* `dns` (string, 可选): 由"nameservers", "domain" 及 "search"构成的映射表.

### 支持的参数

支持以下CNI参数 [CNI_ARGS](https://github.com/containernetworking/cni/blob/master/SPEC.md#parameters):

* `IP`: 需要指定一个以CIDR表示法表示的ip地址，用逗号进行分割
* `GATEWAY`: 需要指定一个网关地址

    (示例: CNI_ARGS="IP=192.168.1.1/24;GATEWAY=192.168.1.254")

同时插件支持能力参数 [capability argument](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md).

* `ips`: 为CNI接口传入多个ip地址

支持以下既定参数 [args conventions](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md#args-in-network-config) :

* `ips` (字符串数组): 用于尝试分配ip的自定义ip列表 (e.g. '192.168.1.1/24')

注意: 如果以上某些参数同时使用，将会通过以下的优先级选择其中一个进行生效

1. 能力参数[capability argument](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md)
1. 既定参数[args conventions](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md#args-in-network-config)
1. CNI参数[CNI_ARGS](https://github.com/containernetworking/cni/blob/master/SPEC.md#parameters)
