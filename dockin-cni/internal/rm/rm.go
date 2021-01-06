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

package rm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/webankfintech/dockin-cni/internal/log"
	"github.com/webankfintech/dockin-cni/internal/model"
)

func GetRMDataByMock(confDir, podName string) (*model.RMData, error) {
	confPath := filepath.Join(confDir, "mock.data")
	content, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, log.Errorf("get mock rm data error, ReadFile, filepath=%v, err=%s",
			confPath, err.Error())
	}
	log.Infof("get mock rm data success: filepath=%v, resp=%s",
		confPath, string(content))

	rmd := &model.RMData{}
	if err := json.Unmarshal(content, rmd); err != nil {
		return nil, log.Errorf("get mock rm data error, Unmarshal, content=%v, err=%s",
			string(content), err.Error())
	}

	if rmd.Code != 0 {
		return nil, log.Errorf("get mock rm data error, code != 0, content=%v",
			string(content))
	}

	return rmd, nil
}

func GetRMDataByPodName(podName, backend string) (*model.RMData, error) {
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	url := fmt.Sprintf("%s?podName=%s", backend, podName)
	log.Infof("get RM data by podName start: podName=%v, url=%v",
		podName, url)
	resp, err := client.Get(url)
	if err != nil {
		return nil, log.Errorf("http get error, RM HTTP GET, url=%v, err=%s",
			url, err.Error())
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, log.Errorf("parse http response error, HTTP.ReadAll, url=%v, err=%s",
			url, err.Error())
	}
	log.Infof("get net info from rm success: url=%v, resp=%s",
		url, string(content))

	rmd := &model.RMData{}
	if err := json.Unmarshal(content, rmd); err != nil {
		return nil, log.Errorf("unmarshal error, Unmarshal, content=%v, err=%s",
			string(content), err.Error())
	}

	if rmd.Code != 0 {
		return nil, log.Errorf("getRMDataByPodName error, code != 0, content=%v",
			string(content))
	}
	return rmd, nil
}
