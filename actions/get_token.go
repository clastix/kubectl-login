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
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type tokenResponse struct {
	IDToken          string `json:"id_token"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
}

type GetToken struct {
	logger *zap.Logger
	tokenEndpoint, OIDClientID, code, pkceCodeVerifier string
}

func NewGetToken(logger *zap.Logger, tokenEndpoint, OIDClientID, code, pkceCodeVerifier string) *GetToken {
	return &GetToken{
		logger: logger,
		tokenEndpoint: tokenEndpoint,
		OIDClientID: OIDClientID,
		code: code,
		pkceCodeVerifier: pkceCodeVerifier,
	}
}

func (r GetToken) Handle() (idToken, refreshToken string, err error) {
	d := url.Values{}
	d.Add("grant_type", "authorization_code")
	d.Add("response_type", "id_token")
	d.Add("client_id", r.OIDClientID)
	d.Add("code", r.code)
	d.Add("code_verifier", r.pkceCodeVerifier)
	d.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")

	var tokenUrl *url.URL
	tokenUrl, err = url.Parse(r.tokenEndpoint)
	if err != nil {
		r.logger.Error("Cannot retrieve OIDC token due to non well-formed endpoint", zap.Error(err), zap.String("tokenEndpoint", r.tokenEndpoint))
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	var res *http.Response
	if res, err = client.Post(tokenUrl.String(), "application/x-www-form-urlencoded", strings.NewReader(d.Encode())); err != nil {
		r.logger.Error("Cannot reach the server", zap.Error(err), zap.String("uri", tokenUrl.String()))
		return
	}
	defer func() { _ = res.Body.Close() }()

	var b []byte
	if b, err = ioutil.ReadAll(res.Body); err != nil {
		r.logger.Error("Cannot read response body", zap.Error(err))
		return
	}
	p := &tokenResponse{}
	if err = json.Unmarshal(b, p); err != nil {
		r.logger.Error("Cannot unmarshal JSON response", zap.Error(err))
		return
	}

	if len(p.Error) > 0 {
		err = errors.New(p.Error)
		r.logger.Error("Token retrieval failed", zap.Error(err))
		return
	}

	return p.IDToken, p.RefreshToken, nil
}
