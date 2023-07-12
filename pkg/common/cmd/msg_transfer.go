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
	"github.com/OpenIMSDK/Open-IM-Server/internal/msgtransfer"
	"github.com/spf13/cobra"
)

// define a message transfer command
type MsgTransferCmd struct {
	*RootCmd
}

// create a message transfer command
func NewMsgTransferCmd() MsgTransferCmd {
	return MsgTransferCmd{NewRootCmd("msgTransfer")}
}

// add rune
func (m *MsgTransferCmd) addRunE() {
	m.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return msgtransfer.StartTransfer(m.getPrometheusPortFlag(cmd))
	}
}

// exec
func (m *MsgTransferCmd) Exec() error {
	m.addRunE()
	return m.Execute()
}
