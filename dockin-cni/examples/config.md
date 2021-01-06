```
{
    "cniVersion": "0.3.1",
    "name": "mynet",
    "type": "bridge",
    "bridge": "mynet0",
    "isDefaultGateway": true,
    "forceAddress": false,
    "ipMasq": true,
    "hairpinMode": true,
    "route": [
        { "dst": "192.168.0.0/16", "gw": "192.168.1.1" }
      ],
    "dns": {
        "nameservers" : ["8.8.8.8"],
        "domain": "example.com",
        "search": [ "example.com" ]
    }
}
```

Network configuration reference
- name (string, required): the name of the network.
- type (string, required): "bridge".
- bridge (string, optional): name of the bridge to use/create. Defaults to "cni0".
- isGateway (boolean, optional): assign an IP address to the bridge. Defaults to false.
- isDefaultGateway (boolean, optional): Sets isGateway to true and makes the assigned IP the default route. Defaults to false.
- forceAddress (boolean, optional): Indicates if a new IP address should be set if the previous value has been changed. Defaults to false.
- ipMasq (boolean, optional): set up IP Masquerade on the host for traffic originating from this network and destined outside of it. Defaults to false.
- mtu (integer, optional): explicitly set MTU to the specified value. Defaults to the value chosen by the kernel.
- hairpinMode (boolean, optional): set hairpin mode for interfaces on the bridge. Defaults to false.
- promiscMode (boolean, optional): set promiscuous mode on the bridge. Defaults to false.
- vlan (int, optional): assign VLAN tag. Defaults to none.
- routes (string, optional): list of routes add to the container namespace. Each route is a dictionary with "dst" and optional "gw" fields. If "gw" is omitted, value of "gateway" will be used.
- dns (string, optional): the dictionary with "nameservers", "domain" and "search".