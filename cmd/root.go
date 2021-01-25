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
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "login",
	Short: "CLI utility to discover and securely login Kubernetes clusters across multiple operating environments",
	Long: `kubectl-login is a CLI utility to discover and securely login Kubernetes clusters across multiple operating
environments, including local setups and cloud providers, i.e. EKS, AKS, GKE.

Based on the configured authentication mechanism (e.g. TLS client, OIDC), it will login users in the Kubernetes clusters
they are allowed to access and generate a kubeconfig for a chosen cluster.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		// Defining the logging options
		logger, err = zap.NewDevelopment()
		if ok, _ := cmd.PersistentFlags().GetBool("verbose"); ok {
			logger, err = zap.NewDevelopment()
		}
		if err != nil {
			return
		}

		if apiServer, _ := cmd.Flags().GetString("api-server"); len(apiServer) != 0 {
			viper.Set(ApiServer, apiServer)
		}
		if oidcServer, _ := cmd.Flags().GetString("oidc-server"); len(oidcServer) != 0 {
			viper.Set(OIDCServer, oidcServer)
		}
		if oidcClientID, _ := cmd.Flags().GetString("oidc-client-id"); len(oidcClientID) != 0 {
			viper.Set(OIDCClientID, oidcClientID)
		}

		if len(viper.GetString(ApiServer)) == 0 {
			err = fmt.Errorf("missing api-server flag value")
		}
		if len(viper.GetString(OIDCServer)) == 0 {
			err = fmt.Errorf("missing oidc-server-url flag value")
		}
		if len(viper.GetString(OIDCClientID)) == 0 {
			err = fmt.Errorf("missing oidc-client-id flag value")
		}

		return
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubectl-login.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Toggle the verbose logging")

	// oidc-insecure-skip-tls-verify
	// oidc-server-ca-path

	rootCmd.PersistentFlags().String("api-server", "", "The Kubernetes API server to connect to")
	// k8s-insecure-skip-tls-verify
	// k8s-ca-path

	rootCmd.PersistentFlags().String("oidc-server", "", "The OIDC server URL to connect to")
	rootCmd.PersistentFlags().String("oidc-client-id", "", "The OIDC client ID provided")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kubectl-login" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kubectl-login")

		p := path.Join(home, ".kubectl-login.yaml")
		if ok, _ := afero.Exists(afero.NewOsFs(), p); !ok {
			_ = afero.WriteFile(afero.NewOsFs(), p, []byte{}, os.ModePerm)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}
}
