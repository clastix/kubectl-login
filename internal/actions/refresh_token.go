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

type RefreshToken struct {
	logger                                      *zap.Logger
	client                                      *http.Client
	refreshEndpoint, oidcClientID, refreshToken string
}

func NewRefreshToken(logger *zap.Logger, insecureSkipVerify bool, refreshEndpoint string, oidcClientID string, refreshToken string) *RefreshToken {
	return &RefreshToken{
		logger:          logger,
		oidcClientID:    oidcClientID,
		refreshEndpoint: refreshEndpoint,
		refreshToken:    refreshToken,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					//nolint:gosec
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
		},
	}
}

func (r RefreshToken) Handle() (idToken, refresh string, err error) {
	d := url.Values{}
	d.Add("grant_type", "refresh_token")
	d.Add("refresh_token", r.refreshToken)
	d.Add("client_id", r.oidcClientID)

	var tokenURL *url.URL
	tokenURL, err = url.Parse(r.refreshEndpoint)
	if err != nil {
		r.logger.Error("Cannot retrieve OIDC token due to non well-formed endpoint", zap.Error(err), zap.String("refreshEndpoint", r.refreshEndpoint))
		return
	}

	var res *http.Response
	if res, err = r.client.Post(tokenURL.String(), "application/x-www-form-urlencoded", strings.NewReader(d.Encode())); err != nil {
		r.logger.Error("Cannot reach the server", zap.Error(err), zap.String("uri", tokenURL.String()))
		return
	}
	defer func() { _ = res.Body.Close() }()

	var b []byte
	if b, err = ioutil.ReadAll(res.Body); err != nil {
		r.logger.Error("Cannot read response body", zap.Error(err))
		return
	}
	t := &tokenResponse{}
	if err = json.Unmarshal(b, t); err != nil {
		r.logger.Error("Cannot unmarshal JSON response", zap.Error(err))
		return
	}

	if len(t.Error) > 0 {
		err = errors.New(t.Error)
		r.logger.Error("Token retrieval failed", zap.Error(err))
		return
	}

	return t.IDToken, t.RefreshToken, nil
}
