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

package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/webankfintech/dockin-cni/internal/log"
	"github.com/webankfintech/dockin-cni/internal/utils"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types"
	cnitypes "github.com/containernetworking/cni/pkg/types"
)

var (
	_defaultIPAMType = "dockin-ipam"
	_defaultIPAMName = "dockin-ipam"
)

type DelegateNetConf struct {
	Conf types.NetConf
	Bytes []byte
}

func NewDelegateNetConf(netInfo *NetInfo, confDir string) (*DelegateNetConf, error) {
	log.Infof("create DelegateNetConf, Type=%s, confDir=%s", netInfo.Type, confDir)

	files, err := libcni.ConfFiles(confDir, []string{".conf", ".json", ".conflist"})
	switch {
	case err != nil:
		return nil, log.Errorf("no networks files found in %s", confDir)
	case len(files) == 0:
		return nil, log.Errorf("confDir %s dir is empty or no conf file type", confDir)
	}

	var (
		bytes []byte
		found = false
		netc  *libcni.NetworkConfig
	)
	for _, confFile := range files {
		bytes, err = ioutil.ReadFile(confFile)
		if err != nil {
			log.Errorf("error read conf file %s: %s", confFile, err)
			continue
		}
		netc, err = libcni.ConfFromBytes(bytes)
		if err != nil {
			log.Errorf("error marshal to network config, conf file %s: %s", confFile, err)
			continue
		}
		if netc.Network.Name == netInfo.Type {
			found = true
			break
		}
	}
	if !found {
		return nil, log.Errorf("no config file exist, name=%s, confDir=%s", netInfo.Type, confDir)
	}
	log.Infof("load delegate net conf success, name=%s, content=%s", netInfo.Type, string(bytes))

	confMap := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &confMap); err != nil {
		return nil, fmt.Errorf("error unmarshal delegate net conf to conf map, type=%s: err=%v",
			netInfo.Type, err)
	}
	bridgeNetConf := &BridgeNetConf{}
	if err := json.Unmarshal(bytes, bridgeNetConf); err != nil {
		return nil, fmt.Errorf("error unmarshal delegate net conf to BridgeNetConf, type=%s: err=%v",
			netInfo.Type, err)
	}

	mask := utils.ParseIPv4Mask(netInfo.SubnetMask)
	size, _ := mask.Size()
	cidrIp := fmt.Sprintf("%s/%d", netInfo.PodIP, size)

	if _, _, err := net.ParseCIDR(cidrIp); err != nil {
		return nil, log.Errorf("failed to parse CIDR %q", netInfo.PodIP)
	}
		ipam := &DockinIPAMConfig{}
	ipam.Type = _defaultIPAMType
	ipam.Name = _defaultIPAMName
	ipam.Addresses = []Address{{
		AddressStr: cidrIp,
		Gateway:    net.ParseIP(netInfo.Gateway),
	}}
		if bridgeNetConf.Routes == nil {
		ipam.Routes = []*cnitypes.Route{
			{
				Dst: func() net.IPNet {
					n, _ := utils.ParseIPNet("0.0.0.0/0")
					return *n
				}(),
				GW: net.ParseIP(netInfo.Gateway),
			},
		}
	} else {
		ipam.Routes = bridgeNetConf.Routes
	}
		if bridgeNetConf.DNS != nil {
		ipam.DNS = bridgeNetConf.DNS
	}

	confMap["ipam"] = ipam

	if bytes, err = json.Marshal(confMap); err != nil {
		return nil, log.Errorf("error marshal conf map to bytes, type %s: err=%s", netInfo.Type, err)
	}

	if err != nil || bytes == nil {
		return nil, log.Errorf("error convert content to NetworkConfig, %s, err=%s", string(bytes), err.Error())
	}
	log.Infof("load conf from file return conf network Type=%s, conf bytes=%s", netInfo.Type, string(bytes))

	delegateConf := &DelegateNetConf{}
	if err := json.Unmarshal(bytes, &delegateConf.Conf); err != nil {
		return nil, log.Errorf("LoadDelegateNetConf: error unmarshal delegate config: %v", err)
	}
	delegateConf.Bytes = bytes
	log.Infof("success create DelegateNetConf, Type=%s, delegate conf=%s", netInfo.Type, string(delegateConf.Bytes))
	return delegateConf, nil
}
