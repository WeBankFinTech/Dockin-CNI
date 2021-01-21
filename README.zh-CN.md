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

## 部署说明
### 1. cni的配置
cni配置文件需要放置到`/etc/cni/net.d`下（kubelet使用的默认配置路径。如果重定向了配置路径，请将配置文件放置到kubelet使用的路径。

配置文件的命名可以参考`00-dockin-cni.json`这样的命名方式

配置文件内容示例：
```
{
    "cniVersion": "0.2.0",
    "name": "dockin-cni",
    "type": "dockin-cni",
    "confDir": "/etc/cni/dockin/net.d",
    "binDir": "/opt/cni/bin",
    "logFile": "/data/kubernetes/dockin-cni.log",
    "logLevel": "debug",
    "backend": "http://localhost:10002/rmController/getPodMultiNetwork"
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
- backend, webhook的访问地址，这里使用dockin-RM的地址作为示例

### 2. 配置Network
这里同样需要创建配置network相关的配置文件。

#### Step1：使用webhook获取network类型
首先，您需要有一个web服务器提供webhook，用于获取pod的网络信息（包括单/双网卡），该web服务需要实现一个带有`podName`url参数的API。比如：
```
<IP>:<port>/<URL>?podName=
```
这里的话，可以使用Dockin-RM作为例子。可以使用`curl`命令来访问rm的以下API：
```
curl 127.0.0.1:10002/rmController/getPodMultiNetwork?podName=<your_pod_name>
```
如果没有出现错误的话，您将会得到如下格式的响应：
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

**这里我们需要关注的是其中`type`字段. 在这个例子中，一共有两个类型: `test` 和 `dockin`**

#### Step2: 创建network配置文件
network 配置通过网桥进行管理，更多细节可参考以下链接：
>https://github.com/containernetworking/plugins/tree/master/plugins/main/bridge

网络配置通过json文件进行存储，存放在`confDir`中（在cni配置中），并且将会传给kubelet创建网络。

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

**上面对配置的内容做了简单介绍，现在来创建network配置文件.**
- 首先，先创建配置文件目录：

目录的路径可以从前面的cni配置文件中的`confDir`找到
```
mkdir -p /etc/cni/dockin/net.d
```

- 然后，创建配置文件

在上面的例子中，我们需要创建两个network配置文件
1.为类型`test`创建配置文件
```
touch /etc/cni/dockin/net.d/test.json
```
配置文件内容：
```JSON
{
  "cniVersion": "0.2.0",
  "name": "test", // type
  "type": "bridge",
  "bridge": "br0"
}
```

2.为类型`dockin`创建配置文件
```
touch /etc/cni/dockin/net.d/dockin.json
```
配置文件内容：
```JSON
{
  "cniVersion": "0.2.0",
  "name": "dockin", // type
  "type": "bridge",
  "bridge": "br0"
}
```

### 3. 将可执行文件放到`binDir`中
你可以使用`make`命令完成`dockin-cni`和`dockin-ipam`的编译，然后将两个可执行文件放到cni配置文件中`binDir`指向的路径。通常情况下，cni的bin目录为`/opt/cni/bin`。

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
