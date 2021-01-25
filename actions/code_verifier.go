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
	"encoding/base64"
	"regexp"

	"go.uber.org/zap"
)

type CodeVerifier struct {
	logger *zap.Logger
}

func NewCodeVerifier(logger *zap.Logger) *CodeVerifier {
	return &CodeVerifier{logger: logger}
}

// Length of code_verifier should be no less than 43 characters and no more than 128 characters, and Base64URL encoded.
func (r CodeVerifier) Handle() (out string, err error) {
	r.logger.Info("Generating PKCE Code Verifier and Challenge")

	b := make([]byte, 50)
	if _, err = rand.Read(b); err != nil {
		r.logger.Error("Cannot read random generate data", zap.Error(err))
		return
	}

	a := base64.URLEncoding.EncodeToString(b)
	regex := regexp.MustCompile(`[^a-zA-Z0-9]`)

	return regex.ReplaceAllString(a, ""), nil
}
