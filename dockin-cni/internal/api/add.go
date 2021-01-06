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
	"github.com/webankfintech/dockin-cni/internal/log"
	"github.com/webankfintech/dockin-cni/internal/model"
	"github.com/webankfintech/dockin-cni/internal/rm"

	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/skel"

	cnitypes "github.com/containernetworking/cni/pkg/types"
)

type DockinCNIAdder struct {
	podName     string
	namespace   string
	containerID string
	cmdArgs     *skel.CmdArgs
	exec        invoke.Exec
	delegates   []*Delegate
}

func (c *DockinCNIAdder) add() (cnitypes.Result, error) {
	if err := c.setPodAndNS(); err != nil {
		return nil, err
	}

	netconf, err := model.LoadNetConf(c.cmdArgs.StdinData)
	if err != nil {
		return nil, log.Errorf("failed to load netconf from stdin=%s, err=%s",
			string(c.cmdArgs.StdinData), err.Error())
	}
	rd, err := rm.GetRMDataByPodName(c.podName, netconf.Backend)
	if err != nil {
		return nil, log.Errorf("get rm data failed, %v", err)
	}
	log.Infof("get pod network information from rm success, podName=%s, %v", c.podName, rd)

	for _, netinfo := range rd.Data {
		d := &Delegate{
			netInfo:     netinfo,
			confDir:     netconf.ConfDir,
			binDir:      netconf.BinDir,
			cmdArgs:     c.cmdArgs,
			podName:     c.podName,
			namespace:   c.namespace,
			containerID: c.containerID,
			exec:        c.exec,
		}
		if err := d.Initialize(); err != nil {
			return nil, log.Errorf("initialize delegate error %s", err.Error())
		}

		log.Infof("success add delegate, netInfo=%v", netinfo)
		c.delegates = append(c.delegates, d)
	}

	var finalResult cnitypes.Result
	for _, delg := range c.delegates {
		delgResult, err := delg.add()
		if err != nil {
			return nil, log.Errorf("exec delegate add error %s", err.Error())
		}
		if delg.netInfo.Master && delgResult != nil {
			finalResult = delgResult
		}
	}

	log.Infof("success add pod network, netInfo %v, result=%s", rd.Data, finalResult.String())
	return finalResult, nil
}

func (c *DockinCNIAdder) setPodAndNS() error {
	k8sArgs := &model.K8sArgs{}
	log.Infof("set k8s pod and namespace, args=%s", c.cmdArgs.Args)
	err := cnitypes.LoadArgs(c.cmdArgs.Args, k8sArgs)
	if err != nil {
		return log.Errorf("failed to load k8s args, as err=%s", err.Error())
	}

	c.podName = string(k8sArgs.K8S_POD_NAME)
	c.namespace = string(k8sArgs.K8S_POD_NAMESPACE)
	c.containerID = string(k8sArgs.K8S_POD_INFRA_CONTAINER_ID)

	log.Infof("set k8s pod info, podName=%s, namespace=%s, contaienrID=%s",
		c.podName, c.namespace, c.containerID)
	return nil
}

func CmdAdd(args *skel.CmdArgs, exec invoke.Exec) (cnitypes.Result, error) {
	log.Infof("add network, input args, ContainerID=%s, Netns=%s, ifName=%s, args=%s, path=%s, stdin=%s",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, string(args.StdinData))

	adder := DockinCNIAdder{
		cmdArgs: args,
		exec:    exec,
	}
	result, err := adder.add()
	if err != nil {
		return nil, log.Errorf("execute add error %s", err.Error())
	}

	log.Infof("success add network, input args, ContainerID=%s, Netns=%s, ifName=%s, args=%s, path=%s, stdin=%s",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, string(args.StdinData))
	return result, nil
}
