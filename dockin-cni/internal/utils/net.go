/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/webankfintech/dockin-cni/internal/log"

	"github.com/containernetworking/cni/pkg/skel"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

func DeleteDefaultGW(args *skel.CmdArgs, ifName string, res *cnitypes.Result) (*current.Result, error) {
	result, err := current.NewResultFromResult(*res)
	if err != nil {
		return nil, log.Errorf("delete default gateway error %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return nil, log.Errorf("error getting netns %v", err)
	}
	defer netns.Close()

	err = netns.Do(func(_ ns.NetNS) error {
		var err error
		link, _ := netlink.LinkByName(ifName)
		routes, _ := netlink.RouteList(link, netlink.FAMILY_ALL)
		for _, nlroute := range routes {
			if nlroute.Dst == nil {
				err = netlink.RouteDel(&nlroute)
			}
		}
		return err
	})
	var newRoutes []*cnitypes.Route
	for _, route := range result.Routes {
		if mask, _ := route.Dst.Mask.Size(); mask != 0 {
			newRoutes = append(newRoutes, route)
		}
	}
	result.Routes = newRoutes
	return result, err
}

func SetDefaultGW(args *skel.CmdArgs, ifName string, gateways []net.IP, res *cnitypes.Result) (*current.Result, error) {

		result, err := current.NewResultFromResult(*res)
	if err != nil {
		return nil, log.Errorf("set default gateway error %v", err)
	}

		netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return nil, log.Errorf("error getting netns %v", err)
	}
	defer netns.Close()

	var newResultDefaultRoutes []*cnitypes.Route

		err = netns.Do(func(_ ns.NetNS) error {
		var err error

				link, _ := netlink.LinkByName(ifName)

				for _, gw := range gateways {

						log.Infof("adding default route on %v (index: %v) to %v", ifName, link.Attrs().Index, gw)
			newDefaultRoute := netlink.Route{
				LinkIndex: link.Attrs().Index,
				Gw:        gw,
			}

			
						_, dstipnet, _ := net.ParseCIDR("::0/0")
			if strings.Count(gw.String(), ":") < 2 {
				_, dstipnet, _ = net.ParseCIDR("0.0.0.0/0")
			}
			newResultDefaultRoutes = append(newResultDefaultRoutes, &cnitypes.Route{Dst: *dstipnet, GW: gw})

						err = netlink.RouteAdd(&newDefaultRoute)
			if err != nil {
				log.Errorf("add route route: %v", err)
			}
		}
		return err
	})

	result.Routes = newResultDefaultRoutes
	return result, err

}

func ParseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		if len(s) != 8 {
			return nil
		}
						m := []int{}
		for i := 0; i < 4; i++ {
			b := "0x" + s[2*i:2*i+2]
			d, err := strconv.ParseInt(b, 0, 0)
			if err != nil {
				return nil
			}
			m = append(m, int(d))
		}
		s := fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
		mask = net.ParseIP(s)
		if mask == nil {
			return nil
		}
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}

func ParseIPNet(s string) (*net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	ipNet.IP = ip
	return ipNet, nil
}
