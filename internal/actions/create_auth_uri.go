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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"

	"go.uber.org/zap"
)

type AuthenticationURI struct {
	logger                                   *zap.Logger
	authEndpoint, oidcClientID, codeVerifier string
}

func NewAuthenticationURI(logger *zap.Logger, oidcClientID, codeVerifier string, configuration *OIDCResponse) *AuthenticationURI {
	return &AuthenticationURI{
		logger:       logger,
		authEndpoint: configuration.AuthorizationEndpoint,
		oidcClientID: oidcClientID,
		codeVerifier: codeVerifier,
	}
}

func (r AuthenticationURI) Handle() (authURI string, err error) {
	r.logger.Info("Creating authorization URI")

	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		r.logger.Error("Cannot read random generate data", zap.Error(err))
		return
	}
	b64State := base64.URLEncoding.EncodeToString(b)
	re := regexp.MustCompile(`[\W_]`)
	state := re.ReplaceAllString(b64State, "")

	hash := sha256.Sum256([]byte(r.codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	var u *url.URL
	u, err = url.Parse(r.authEndpoint)
	if err != nil {
		r.logger.Error("Cannot parse auth endpoint", zap.String("authEndpoint", r.authEndpoint), zap.Error(err))
		err = fmt.Errorf("non well-formed endpoint")
		return
	}

	qs := u.Query()
	qs.Set("response_type", "code")
	qs.Set("client_id", r.oidcClientID)
	qs.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	qs.Set("scope", "openid+groups+offline_access")
	qs.Set("state", state)
	qs.Set("prompt", "consent")
	qs.Set("code_challenge", codeChallenge)
	qs.Set("code_challenge_method", codeChallengeMethod)
	qs.Set("access_type", "offline")

	authURL := url.URL{
		Scheme:     u.Scheme,
		Host:       u.Host,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: true,
		RawQuery: func(value string) (out string) {
			out, _ = url.QueryUnescape(value)
			return
		}(qs.Encode()),
	}

	return authURL.String(), nil
}
