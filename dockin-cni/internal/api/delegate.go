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

package api

import (
	"context"
	"os"
	"path/filepath"

	"github.com/webankfintech/dockin-cni/internal/log"
	"github.com/webankfintech/dockin-cni/internal/model"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"

	cnitypes "github.com/containernetworking/cni/pkg/types"
)

type Delegate struct {
	delegateNetConf *model.DelegateNetConf
	netInfo         *model.NetInfo
	confDir         string
	binDir          string
	cmdArgs         *skel.CmdArgs
	podName         string
	namespace       string
	containerID     string
	exec            invoke.Exec
}

func (d *Delegate) Initialize() error {
	delegateNetConf, err := model.NewDelegateNetConf(d.netInfo, d.confDir)
	if err != nil {
		return log.Errorf(err.Error())
	}
	d.delegateNetConf = delegateNetConf
	log.Infof("set delegate conf, netinfo=%v, delegate conf=%s", d, string(delegateNetConf.Bytes))
	return nil
}

func (d *Delegate) add() (cnitypes.Result, error) {
	var (
		result cnitypes.Result
		err    error
	)
	nsname := os.Getenv("CNI_NETNS")
	if err = d.validateIfName(nsname, d.netInfo.IfName); err != nil {
		return nil, log.Errorf(err.Error())
	}

	log.Debugf("prepare to set the CNI_IFNAME=%s", d.netInfo.IfName)
	if os.Setenv("CNI_IFNAME", d.netInfo.IfName) != nil {
		return nil, log.Errorf("failed to set current ifName=%s, podName=%s", d.netInfo.IfName, d.podName)
	}
	binDirs := filepath.SplitList(os.Getenv("CNI_PATH"))
	binDirs = append([]string{d.binDir}, binDirs...)
	cniNet := libcni.NewCNIConfig(binDirs, d.exec)

	log.Infof("prepare to parse NetworkConfig, byte=%s", string(d.delegateNetConf.Bytes))
	networkConfig, err := libcni.ConfFromBytes(d.delegateNetConf.Bytes)
	if err != nil {
		return nil, log.Errorf("error in converting the raw bytes to conf: %v", err)
	}

	log.Infof("prepare call cni api, conf = %s", string(networkConfig.Bytes))
	runtimeConf := d.createDelegateCNIRuntimeConf()
	result, err = cniNet.AddNetwork(context.Background(), networkConfig, runtimeConf)
	if err != nil {
		return nil, log.Errorf("error in getting result from AddNetwork: %v", err)
	}

	log.Infof("success call cni api to create network, podName=%s, result=%s", d.podName, result.String())
	return result, nil
}

func (d *Delegate) createDelegateCNIRuntimeConf() *libcni.RuntimeConf {
	rt := &libcni.RuntimeConf{
		ContainerID: d.cmdArgs.ContainerID,
		NetNS:       d.cmdArgs.Netns,
		IfName:      d.netInfo.IfName,
		Args: [][2]string{
			{"IgnoreUnknown", string("true")},
			{"K8S_POD_NAMESPACE", d.namespace},
			{"K8S_POD_NAME", d.podName},
			{"K8S_POD_INFRA_CONTAINER_ID", d.containerID},
		},
	}

	return rt
}

func (d *Delegate) validateIfName(nsname string, ifname string) error {
	log.Infof("validate ifName, ns=%s, ifName=%s", nsname, ifname)
	podNs, err := ns.GetNS(nsname)
	if err != nil {
		return log.Errorf(err.Error())
	}

	err = podNs.Do(func(_ ns.NetNS) error {
		_, err := netlink.LinkByName(ifname)
		if err != nil {
			if err.Error() == "Link not found" {
				return nil
			}
			return log.Errorf("link by name error %v", err)
		}
		return log.Errorf("interface name %s already exists", ifname)
	})

	log.Infof("validate ifName end, nsname=%s, ifName=%s, err=%v", nsname, ifname, err)
	return err
}

func (d *Delegate) setDefaultGateway() error {
	return nil
}

func (d *Delegate) Delete() {

}
