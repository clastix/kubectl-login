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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	json2 "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/pkg/apis/clientauthentication"

	"github.com/clastix/kubectl-login/actions"
)

var tokenCmd = &cobra.Command{
	Use:           "token",
	Short:         "Return a valid id_token to kubectl",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var idToken string
		if idToken = viper.GetString(IDToken); len(idToken) == 0 {
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

			var idToken, refreshToken string
			idToken, refreshToken, err = actions.NewRefreshToken(logger, true, viper.GetString(TokenEndpoint), viper.GetString(OIDCClientID), viper.GetString(RefreshToken)).Handle()
			if err != nil {
				return fmt.Errorf("cannot refresh token due to an error (%w)", err)
			}

			viper.Set(IDToken, idToken)
			viper.Set(RefreshToken, refreshToken)

		}
		ec := &clientauthentication.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ExecCredential",
				APIVersion: "client.authentication.k8s.io/v1beta1",
			},
			Status: &clientauthentication.ExecCredentialStatus{
				Token: viper.GetString(IDToken),
			},
		}

		scheme := runtime.NewScheme()
		encoder := json2.NewSerializerWithOptions(json2.SimpleMetaFactory{}, scheme, scheme, json2.SerializerOptions{
			Strict: true,
		})
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
