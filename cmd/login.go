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
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/clastix/kubectl-login/actions"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Perform the authentication login to the desired OIDC server",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		logger.Info("Starting the login procedure")

		var authResponse *actions.OIDCResponse
		if authResponse, err = actions.NewOIDCConfiguration(logger, true).Handle(viper.GetString(OIDCServer)); err != nil {
			return err
		}

		var codeVerifier string
		if codeVerifier, err = actions.NewCodeVerifier(logger).Handle(); err != nil {
			return err
		}

		var url string
		if url, err = actions.NewAuthenticationURI(logger, viper.GetString(OIDCClientID), codeVerifier, authResponse).Handle(); err != nil {
			return err
		}

		fmt.Println("")
		fmt.Println("Proceed to login to the following link using your browser:")
		fmt.Println("")
		fmt.Println(url)
		fmt.Println("")

		prompt := promptui.Prompt{
			Label: "Enter verification code",
			Validate: func(s string) (err error) {
				if len(s) == 0 {
					err = fmt.Errorf("an empty string is not a valid code")
				}
				return
			},
		}
		var code string
		if code, err = prompt.Run(); err != nil {
			return err
		}

		var idToken, refreshToken string
		idToken, refreshToken, err = actions.NewGetToken(logger, authResponse.TokenEndpoint, viper.GetString(OIDCClientID), code, codeVerifier).Handle()
		if err != nil {
			return fmt.Errorf("Cannot proceed to login due to an error: %w", err)
		}

		//viper.Set("codeChallengeMethod", "S256")
		//viper.Set("authorizationEndpoint", authResponse.AuthorizationEndpoint)
		viper.Set(TokenEndpoint, authResponse.TokenEndpoint)
		//viper.Set("introspectionEndpoint", authResponse.IntrospectionEndpoint)
		//viper.Set("userInfoEndpoint", authResponse.UserInfoEndpoint)
		//viper.Set("endSessionEndpoint", authResponse.EndSessionEndpoint)
		viper.Set(IDToken, idToken)
		viper.Set(RefreshToken, refreshToken)

		return viper.WriteConfig()
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
