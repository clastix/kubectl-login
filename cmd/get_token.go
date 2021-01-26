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

package cmd

import (
	"bytes"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	clientauthenticationv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"

	"github.com/clastix/kubectl-login/internal/actions"
)

var tokenCmd = &cobra.Command{
	Use:   "get-token",
	Short: "Return a credential execution required by kubectl with the updated ID token",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var idToken string
		if idToken = viper.GetString(TokenID); len(idToken) == 0 {
			return fmt.Errorf("the ID Token is not yet configured, please issue the login process first")
		}

		logger.Info("Decoding the ID token as JWT")
		claims := &jwt.MapClaims{}
		parser := jwt.Parser{SkipClaimsValidation: true}
		if _, _, err = parser.ParseUnverified(idToken, claims); err != nil {
			return fmt.Errorf("token ID is a non JWT encoded string (%w)", err)
		}

		if err = claims.Valid(); err != nil {
			logger.Info("proceeding to token refresh")
			logger.Debug("JWT claim is not valid due to error", zap.Error(err))

			var refreshToken string
			idToken, refreshToken, err = actions.NewRefreshToken(logger, true, viper.GetString(TokenEndpoint), viper.GetString(OIDCClientID), viper.GetString(TokenRefresh)).Handle()
			if err != nil {
				return fmt.Errorf("cannot refresh token due to an error (%w)", err)
			}

			viper.Set(TokenID, idToken)
			viper.Set(TokenRefresh, refreshToken)

			defer func() {
				if err = viper.WriteConfig(); err != nil {
					logger.Error("Cannot write configuration file", zap.Error(err))
				}
			}()
		}
		ec := &clientauthenticationv1beta1.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ExecCredential",
				APIVersion: "client.authentication.k8s.io/v1beta1",
			},
			Status: &clientauthenticationv1beta1.ExecCredentialStatus{
				Token: viper.GetString(TokenID),
			},
		}

		scheme := runtime.NewScheme()
		encoder := json.NewSerializerWithOptions(json.SimpleMetaFactory{}, scheme, scheme, json.SerializerOptions{})
		a := bytes.NewBuffer([]byte{})
		if err = encoder.Encode(ec, a); err != nil {
			return fmt.Errorf("cannot encode kubeconfig to JSON (%w)", err)
		}

		fmt.Println(a.String())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}
