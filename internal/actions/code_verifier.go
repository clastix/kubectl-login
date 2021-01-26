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
	"math/rand"

	"go.uber.org/zap"
)

const dictBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

type CodeVerifier struct {
	logger *zap.Logger
}

func NewCodeVerifier(logger *zap.Logger) *CodeVerifier {
	return &CodeVerifier{logger: logger}
}

func (r CodeVerifier) Handle() (out string) {
	b := make([]byte, 128)

	r.logger.Info("Generating PKCE Code Verifier and Challenge")
	defer func() { r.logger.Info("PKCE code verifier generated", zap.ByteString("code", b)) }()

	for i := range b {
		b[i] = dictBytes[rand.Intn(len(dictBytes))]
	}

	return string(b)
}
