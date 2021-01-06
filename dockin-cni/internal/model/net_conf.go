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

	"github.com/webankfintech/dockin-cni/internal/log"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

const (
	defaultConfDir = "/etc/cni/dockin/net.d"
	defaultBinDir  = "/opt/cni/bin"
)

type NetConf struct {
	types.NetConf

			RawPrevResult *map[string]interface{} `json:"prevResult"`
	PrevResult    *current.Result         `json:"-"`

	ConfDir  string `json:"confDir"`
	BinDir   string `json:"binDir"`
	LogFile  string `json:"logFile"`
	LogLevel string `json:"logLevel"`

		Backend       string         `json:"backend"`
	RuntimeConfig *RuntimeConfig `json:"runtimeConfig,omitempty"`
}

func (n *NetConf) ToString() string {
	d, _ := json.Marshal(n)
	return string(d)
}

type RuntimeConfig struct {
	PortMaps  []*PortMapEntry `json:"portMappings,omitempty"`
	Bandwidth *BandwidthEntry `json:"bandwidth,omitempty"`
	IPs       []string        `json:"ips,omitempty"`
	Mac       string          `json:"mac,omitempty"`
}

type PortMapEntry struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

type BandwidthEntry struct {
	IngressRate  int `json:"ingressRate"`
	IngressBurst int `json:"ingressBurst"`

	EgressRate  int `json:"egressRate"`
	EgressBurst int `json:"egressBurst"`
}

func LoadNetConf(bytes []byte) (*NetConf, error) {
	netconf := &NetConf{}

	log.Infof("load net conf, input bytes %s", string(bytes))
	if err := json.Unmarshal(bytes, netconf); err != nil {
		return nil, log.Errorf("failed to load net conf: %v", err)
	}

		if netconf.LogFile != "" {
		log.SetLogFile(netconf.LogFile)
	}
		if netconf.LogLevel != "" {
		log.SetLogLevel(netconf.LogLevel)
	}

		if netconf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(netconf.RawPrevResult)
		if err != nil {
			return nil, log.Errorf("could not serialize prevResult: %v", err)
		}
		res, err := version.NewResult(netconf.CNIVersion, resultBytes)
		if err != nil {
			return nil, log.Errorf("could not parse prevResult: %v", err)
		}
		netconf.RawPrevResult = nil
		netconf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, log.Errorf("could not convert result to current version: %v", err)
		}
	}

	if netconf.ConfDir == "" {
		netconf.ConfDir = defaultConfDir
	}

	if netconf.BinDir == "" {
		netconf.BinDir = defaultBinDir
	}

	return netconf, nil
}
