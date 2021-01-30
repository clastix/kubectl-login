/*
Copyright © 2021 Clastix Labs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/clastix/kubectl-login/internal/oidc"
)

const (
	codeChallengeMethod = "S256"
)

type OIDCResponse struct {
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	IntrospectionEndpoint         string   `json:"introspection_endpoint"`
	UserInfoEndpoint              string   `json:"userinfo_endpoint"`
	EndSessionEndpoint            string   `json:"end_session_endpoint"`
	GrantTypesSupported           []string `json:"grant_types_supported"`
	ResponseTypesSupported        []string `json:"response_types_supported"`
	ResponseModesSupported        []string `json:"response_modes_supported"`
	ClaimsSupported               []string `json:"claims_supported"`
	ScopesSupported               []string `json:"scopes_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	Error                         string   `json:"error"`
}

type PKCELogin struct {
	logger *zap.Logger
	client *oidc.HTTPClient
}

func NewOIDCConfiguration(logger *zap.Logger, oidcClient *oidc.HTTPClient) *PKCELogin {
	return &PKCELogin{
		logger: logger,
		client: oidcClient,
	}
}

func (r PKCELogin) Handle(oidcServer string) (response *OIDCResponse, err error) {
	r.logger.Info("Getting OIDC configuration from the server", zap.String("OIDCServer", oidcServer))

	var res *http.Response
	res, err = r.client.Get(fmt.Sprintf("%s/.well-known/openid-configuration", oidcServer))
	if err != nil {
		r.logger.Error("the server returned an error", zap.String("OIDCServer", oidcServer), zap.Error(err))
		err = fmt.Errorf("the server returned an error")
		return
	}
	defer func() { _ = res.Body.Close() }()

	response = &OIDCResponse{}
	b, _ := ioutil.ReadAll(res.Body)
	if err = json.Unmarshal(b, response); err != nil {
		r.logger.Error("Cannot unmarshal OIDC configuration", zap.String("OIDCServer", oidcServer), zap.Error(err), zap.ByteString("body", b))
		err = fmt.Errorf("the response body is not a valid JSON")
		return
	}
	if len(response.Error) > 0 {
		err = fmt.Errorf("server returned the error %s", response.Error)
	}

	return
}
