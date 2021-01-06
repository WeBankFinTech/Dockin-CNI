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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/webankfintech/dockin-cni/internal/api"

	"github.com/containernetworking/cni/pkg/skel"
	cniversion "github.com/containernetworking/cni/pkg/version"
)

var version = "1.0.0"

func printVersion() string {
	return fmt.Sprintf("dockin-cni used to create pod network informations, "+
		"interact with resource manager, current version:%s", version)
}

func main() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	ver := false
	flag.BoolVar(&ver, "v", false, "print dockin cni version")
	flag.Parse()
	if ver == true {
		fmt.Printf("%s\n", printVersion())
		return
	}

	skel.PluginMain(
		func(args *skel.CmdArgs) error {
			result, err := api.CmdAdd(args, nil)
			if err != nil {
				return err
			}
			return result.Print()
		},
		func(args *skel.CmdArgs) error {
			result, err := api.Check(args, nil)
			if err != nil {
				return err
			}
			return result.Print()
		},
		func(args *skel.CmdArgs) error {
			return api.CmdDelete(args, nil)
		},

		cniversion.All, printVersion(),
	)
}
