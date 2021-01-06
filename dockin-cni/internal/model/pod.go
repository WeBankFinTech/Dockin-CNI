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
	"errors"
	"fmt"
)

type PodInfo struct {
	Name string			`json:"name"`
	Namespace string	`json:"namespace"`
	UID string			`json:"uid"`
	NetInfos []*NetInfo	`json:"netinfo"`
}

func (p *PodInfo) GetNetInfoByName(name string) (*NetInfo, error){
	for _, v := range p.NetInfos {
		if name == v.Type {
			return v, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("GetNetInfoByName: no net info for name=%s", name))
}

func (p *PodInfo)ToString() string {
	d, _ := json.Marshal(p)
	return string(d)
}

type RMData struct {
	Code int			`json:"code"`
	ReqID string		`json:"reqId"`
	Message string 		`json:"message"`
	Data []*NetInfo		`json:"data"`
}

type NetInfo struct {
	Type string 		`json:"type"`
	PodIP string 		`json:"podIp"`
	IfName string		`json:"ifName"`
	SubnetMask string 	`json:"subnetMask"`
	Gateway string 		`json:"gateway"`
	Master bool         `json:"master"`
}

