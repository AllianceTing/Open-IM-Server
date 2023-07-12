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

package api

import (
	"github.com/OpenIMSDK/Open-IM-Server/pkg/a2r"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/discoveryregistry"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/auth"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/rpcclient"
	"github.com/gin-gonic/gin"
)

// export AuthApi
type AuthApi rpcclient.Auth

// export NewAuthAPI
func NewAuthAPI(discov discoveryregistry.SvcDiscoveryRegistry) AuthApi {
	return AuthApi(*rpcclient.NewAuth(discov))
}

// export AuthAPI to call UserToken
func (o *AuthApi) UserToken(c *gin.Context) {
	a2r.Call(auth.AuthClient.UserToken, o.Client, c)
}

// export AuthAPI to parse Token
func (o *AuthApi) ParseToken(c *gin.Context) {
	a2r.Call(auth.AuthClient.ParseToken, o.Client, c)
}

// export AuthAPI to force logout
func (o *AuthApi) ForceLogout(c *gin.Context) {
	a2r.Call(auth.AuthClient.ForceLogout, o.Client, c)
}
