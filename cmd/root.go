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
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/clastix/kubectl-login/internal/actions"
	"github.com/clastix/kubectl-login/internal/oidc"
)

var cfgFile string

var defaultKubeConfigPath = func() string {
	if c := os.Getenv("KUBECONFIG"); len(c) > 0 {
		return c
	}
	home, _ := homedir.Dir()
	p := path.Join(home, ".kube", "config")
	if ok, _ := afero.Exists(afero.NewOsFs(), p); !ok {
		_ = afero.WriteFile(afero.NewOsFs(), p, []byte{}, 0600)
	}
	return p
}

var rootCmd = &cobra.Command{
	Use:   "login",
	Short: "CLI utility to discover and securely login Kubernetes clusters across multiple operating environments",
	Long: `kubectl-login is a CLI utility to discover and securely login Kubernetes clusters across multiple operating
environments, including local setups and cloud providers, i.e. EKS, AKS, GKE.

Based on the configured authentication mechanism (e.g. TLS client, OIDC), it will login users in the Kubernetes clusters
they are allowed to access and generate a kubeconfig for a chosen cluster.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if ok, _ := cmd.Flags().GetBool("verbose"); ok {
			logger, err = zap.NewDevelopment()
			if err != nil {
				return
			}
		}

		if v, _ := cmd.Flags().GetString(flagsMap[OIDCServer]); len(v) > 0 {
			viper.Set(OIDCServer, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[OIDCClientID]); len(v) > 0 {
			viper.Set(OIDCClientID, v)
		}

		if cmd.Flag(flagsMap[OIDCTimeoutDuration]).Changed {
			v, _ := cmd.Flags().GetDuration(flagsMap[OIDCTimeoutDuration])
			viper.Set(OIDCTimeoutDuration, v)
		}
		if cmd.Flag(flagsMap[OIDCSkipTLSVerify]).Changed {
			v, _ := cmd.Flags().GetBool(flagsMap[OIDCSkipTLSVerify])
			viper.Set(OIDCSkipTLSVerify, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[OIDCCertificateAuthority]); len(v) > 0 {
			viper.Set(OIDCCertificateAuthority, v)
		}

		if v, _ := cmd.Flags().GetString(flagsMap[K8SAPIServer]); len(v) > 0 {
			viper.Set(K8SAPIServer, v)
		}
		if cmd.Flag(flagsMap[K8SSkipTLSVerify]).Changed {
			v, _ := cmd.Flags().GetBool(flagsMap[K8SSkipTLSVerify])
			viper.Set(K8SSkipTLSVerify, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[K8SCertificateAuthorityPath]); len(v) > 0 {
			viper.Set(K8SCertificateAuthorityPath, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[KubeconfigPath]); len(v) > 0 {
			viper.Set(KubeconfigPath, v)
		}

		if v := viper.GetString(OIDCServer); len(v) == 0 {
			return errors.New("missing OIDC server endpoint")
		}
		if v := viper.GetString(OIDCClientID); len(v) == 0 {
			return errors.New("missing OIDC server endpoint")
		}
		if v := viper.GetString(K8SAPIServer); len(v) == 0 {
			return errors.New("missing Kubernetes API server")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		logger.Info("Starting the login procedure")

		// Creating OIDC server HTTP client with TLS handling
		var client *oidc.HTTPClient
		client, err = oidc.NewHTTPClient(viper.GetString(OIDCCertificateAuthority), viper.GetDuration(OIDCTimeoutDuration), viper.GetBool(OIDCSkipTLSVerify))
		if err != nil {
			return
		}

		// Gathering the OIDC server configuration
		var res *actions.OIDCResponse
		if res, err = actions.NewOIDCConfiguration(logger, client).Handle(viper.GetString(OIDCServer)); err != nil {
			return fmt.Errorf("cannot obtain the OIDC configuration (%w)", err)
		}

		pkce := actions.NewCodeVerifier(logger).Handle()

		var loginURL string
		loginURL, err = actions.NewAuthenticationURI(logger, viper.GetString(OIDCClientID), pkce, res).Handle()
		if err != nil {
			return fmt.Errorf("cannot generate the authentatication URI (%w)", err)
		}

		fmt.Println("")
		fmt.Println("Proceed to login to the following link using your browser:")
		fmt.Println("")
		fmt.Println(loginURL)
		fmt.Println("")

		var code string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Type the verification code: ")
		code, _ = reader.ReadString('\n')
		code = strings.TrimSuffix(code, "\n")

		logger.Debug("User input code is " + code)

		var token, refresh string
		token, refresh, err = actions.NewGetToken(logger, res.TokenEndpoint, viper.GetString(OIDCClientID), code, pkce, client).Handle()
		if err != nil {
			return fmt.Errorf("cannot proceed to login due to an error (%w)", err)
		}

		viper.Set(TokenEndpoint, res.TokenEndpoint)
		viper.Set(TokenID, token)
		viper.Set(TokenRefresh, refresh)

		defer func() {
			if err = viper.WriteConfig(); err != nil {
				logger.Error("Cannot write configuration file", zap.Error(err))
			}
		}()

		var p string
		if p = viper.GetString(KubeconfigPath); len(p) == 0 {
			p = defaultKubeConfigPath()
		}
		var cfg *clientcmdapi.Config
		var cfgErr error
		if cfg, cfgErr = clientcmd.LoadFromFile(p); cfgErr != nil {
			cfg, _ = clientcmd.Load(nil)
		}

		u, _ := url.Parse(viper.GetString(K8SAPIServer))
		name := strings.Join([]string{u.Scheme, u.Hostname(), u.Port()}, "_")

		cfg.CurrentContext = "oidc"
		cfg.Clusters[name] = &clientcmdapi.Cluster{
			Server:                viper.GetString(K8SAPIServer),
			InsecureSkipTLSVerify: viper.GetBool(K8SSkipTLSVerify),
			CertificateAuthorityData: func() (b []byte) {
				if viper.GetBool(K8SSkipTLSVerify) {
					return nil
				}
				b, err = afero.ReadFile(afero.NewOsFs(), viper.GetString(K8SCertificateAuthorityPath))
				return
			}(),
		}
		cfg.Contexts["oidc"] = &clientcmdapi.Context{
			Cluster:  name,
			AuthInfo: "oidc",
		}
		cfg.AuthInfos["oidc"] = &clientcmdapi.AuthInfo{
			Exec: &clientcmdapi.ExecConfig{
				Command:    "kubectl",
				Args:       []string{"login", tokenCmd.Use},
				APIVersion: "client.authentication.k8s.io/v1beta1",
			},
		}
		if err != nil {
			return fmt.Errorf("cannot read Kubernetes CA from file (%w)", err)
		}

		if err = clientcmd.WriteToFile(*cfg, p); err != nil {
			return fmt.Errorf("cannot save generated kubeconfig (%w)", err)
		}

		fmt.Println("Your login procedure has been completed!")
		fmt.Println("")
		fmt.Printf("The Kubernetes configuration file has been merged in your current export KUBECONFIG: %s", p)
		fmt.Println("")

		fmt.Println("")
		fmt.Println("Happy Kubernetes interaction!")

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubectl-login.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Toggle the verbose logging")

	rootCmd.PersistentFlags().String(flagsMap[OIDCServer], viper.GetString(OIDCServer), "The OIDC server URL to connect to")
	rootCmd.PersistentFlags().String(flagsMap[OIDCClientID], viper.GetString(OIDCClientID), "The OIDC client ID provided")
	rootCmd.PersistentFlags().Bool(flagsMap[OIDCSkipTLSVerify], viper.GetBool(OIDCSkipTLSVerify), "Disable TLS certificate verification for the OIDC server")
	rootCmd.PersistentFlags().Duration(flagsMap[OIDCTimeoutDuration], viper.GetDuration(OIDCTimeoutDuration), "Define the timeout in duration for the HTTP requests to the OIDC server")
	rootCmd.PersistentFlags().String(flagsMap[OIDCCertificateAuthority], viper.GetString(OIDCCertificateAuthority), "Path to the OIDC server certificate authority PEM encoded file")

	rootCmd.PersistentFlags().String(flagsMap[K8SAPIServer], viper.GetString(K8SAPIServer), "Endpoint of the Kubernetes API server to connect to")
	rootCmd.PersistentFlags().Bool(flagsMap[K8SSkipTLSVerify], viper.GetBool(K8SSkipTLSVerify), "Disable TLS certificate verification for the Kubernetes API server")
	rootCmd.PersistentFlags().String(flagsMap[K8SCertificateAuthorityPath], viper.GetString(K8SCertificateAuthorityPath), "Path to the Kubernetes API server certificate authority PEM encoded file")

	rootCmd.PersistentFlags().String(flagsMap[KubeconfigPath], "", "Path to the generated kubeconfig file upon resulting login procedure to access the Kubernetes cluster, leave empty for the KUBECONFIG environment variable or default location ($HOME/.kube/config)")
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
