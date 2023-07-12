// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/discoveryregistry"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/startrpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// define a rpccmd struct
type RpcCmd struct {
	*RootCmd
}

// create a new rpc command
func NewRpcCmd(name string) *RpcCmd {
	authCmd := &RpcCmd{NewRootCmd(name)}
	return authCmd
}

// exec
func (a *RpcCmd) Exec() error {
	a.Command.Run = func(cmd *cobra.Command, args []string) {
		a.port = a.getPortFlag(cmd)
		a.prometheusPort = a.getPrometheusPortFlag(cmd)
	}
	return a.Execute()
}

// start server
func (a *RpcCmd) StartSvr(
	name string,
	rpcFn func(client discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error,
) error {
	if a.GetPortFlag() == 0 {
		return errors.New("port is required")
	}
	return startrpc.Start(a.GetPortFlag(), name, a.GetPrometheusPortFlag(), rpcFn)
}
