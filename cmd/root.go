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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/clastix/kubectl-login/internal/actions"
	"github.com/clastix/kubectl-login/internal/oidc"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "login",
	Short: "CLI utility to discover and securely login Kubernetes clusters across multiple operating environments",
	Long: `kubectl-login is a CLI utility to discover and securely login Kubernetes clusters across multiple operating
environments, including local setups and cloud providers, i.e. EKS, AKS, GKE.

Based on the configured authentication mechanism (e.g. TLS client, OIDC), it will login users in the Kubernetes clusters
they are allowed to access and generate a kubeconfig for a chosen cluster.`,
	SilenceUsage:  true,
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
		if v, _ := cmd.Flags().GetBool(flagsMap[OIDCSkipTlsVerify]); true {
			viper.Set(OIDCSkipTlsVerify, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[OIDCCertificateAuthority]); len(v) > 0 {
			viper.Set(OIDCCertificateAuthority, v)
		}

		if v, _ := cmd.Flags().GetString(flagsMap[K8SAPIServer]); len(v) > 0 {
			viper.Set(K8SAPIServer, v)
		}
		if v, _ := cmd.Flags().GetBool(flagsMap[K8SSkipTlsVerify]); true {
			viper.Set(K8SSkipTlsVerify, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[K8SCertificateAuthorityPath]); len(v) > 0 {
			viper.Set(K8SCertificateAuthorityPath, v)
		}
		if v, _ := cmd.Flags().GetString(flagsMap[KubeconfigPath]); len(v) > 0 {
			viper.Set(KubeconfigPath, v)
		}

		if v := viper.GetString(OIDCServer); len(v) == 0 {
			return errors.New("Missing OIDC server endpoint")
		}
		if v := viper.GetString(OIDCClientID); len(v) == 0 {
			return errors.New("Missing OIDC server endpoint")
		}
		if v := viper.GetString(K8SAPIServer); len(v) == 0 {
			return errors.New("Missing Kubernetes API server")
		}
		if v := viper.GetString(KubeconfigPath); len(v) == 0 {
			return errors.New("Missing path to the resulting kubeconfig")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		logger.Info("Starting the login procedure")

		// Creating OIDC server HTTP client with TLS handling
		var client *oidc.HttpClient
		client, err = oidc.NewHTTPClient(viper.GetString(OIDCCertificateAuthority), viper.GetBool(OIDCSkipTlsVerify))
		if err != nil {
			return
		}

		// Gathering the OIDC server configuration
		var res *actions.OIDCResponse
		if res, err = actions.NewOIDCConfiguration(logger, client).Handle(viper.GetString(OIDCServer)); err != nil {
			return err
		}

		pkce := actions.NewCodeVerifier(logger).Handle()

		var url string
		url, err = actions.NewAuthenticationURI(logger, viper.GetString(OIDCClientID), pkce, res).Handle()
		if err != nil {
			return err
		}

		fmt.Println("")
		fmt.Println("Proceed to login to the following link using your browser:")
		fmt.Println("")
		fmt.Println(url)
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
			return fmt.Errorf("Cannot proceed to login due to an error: %w", err)
		}

		viper.Set(TokenEndpoint, res.TokenEndpoint)
		viper.Set(TokenID, token)
		viper.Set(TokenRefresh, refresh)

		defer func() {
			if err = viper.WriteConfig(); err != nil {
				logger.Error("Cannot write configuration file", zap.Error(err))
			}
		}()

		config := &v1.Config{
			Kind:           "Config",
			APIVersion:     "v1",
			CurrentContext: "oidc",
			Clusters: []v1.NamedCluster{
				{
					Name: "kubernetes",
					Cluster: v1.Cluster{
						Server:                viper.GetString(K8SAPIServer),
						InsecureSkipTLSVerify: viper.GetBool(K8SSkipTlsVerify),
						CertificateAuthorityData: func() (b []byte) {
							if viper.GetBool(K8SSkipTlsVerify) {
								return nil
							}
							b, err = afero.ReadFile(afero.NewOsFs(), viper.GetString(K8SCertificateAuthorityPath))
							return
						}(),
					},
				},
			},
			Contexts: []v1.NamedContext{
				{
					Name: "oidc",
					Context: v1.Context{
						Cluster:  "kubernetes",
						AuthInfo: "oidc",
					},
				},
			},
			AuthInfos: []v1.NamedAuthInfo{
				{
					Name: "oidc",
					AuthInfo: v1.AuthInfo{
						Exec: &v1.ExecConfig{
							Command:    "kubectl",
							Args:       []string{"login", tokenCmd.Use},
							APIVersion: "client.authentication.k8s.io/v1beta1",
						},
					},
				},
			},
		}
		if err != nil {
			return fmt.Errorf("Cannot read Kubernetes CA from file: %w", err)
		}

		scheme := runtime.NewScheme()
		e := json.NewSerializerWithOptions(yaml.SimpleMetaFactory{}, scheme, scheme, json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: false,
		})

		a := bytes.NewBuffer([]byte{})
		if err = e.Encode(config, a); err != nil {
			return fmt.Errorf("cannot encode kubeconfig to YAML (%w)", err)
		}

		path, _ := cmd.Flags().GetString("kubeconfig-path")

		if err = afero.WriteFile(afero.NewOsFs(), path, a.Bytes(), os.ModePerm); err != nil {
			msg := "cannot save generated kubeconfig"
			logger.Error(msg, zap.Error(err))
			return errors.New(msg)
		}

		fmt.Println("Your login procedure has been completed!")
		fmt.Println("")
		fmt.Println("You can start interacting with your Kubernetes cluster using the generated kubeconfig file:")
		fmt.Printf("export KUBECONFIG=%s", path)
		fmt.Println("")
		fmt.Println("")
		fmt.Println("Happy Kubernetes interaction!")

		return nil
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().String(flagsMap[OIDCServer], "", "The OIDC server URL to connect to")
	rootCmd.PersistentFlags().String(flagsMap[OIDCClientID], "", "The OIDC client ID provided")
	rootCmd.PersistentFlags().Bool(flagsMap[OIDCSkipTlsVerify], false, "Disable TLS certificate verification for the OIDC server")
	rootCmd.PersistentFlags().String(flagsMap[OIDCCertificateAuthority], "", "Path to the OIDC server certificate authority PEM encoded file")

	rootCmd.PersistentFlags().String(flagsMap[K8SAPIServer], "", "Endpoint of the Kubernetes API server to connect to")
	rootCmd.PersistentFlags().Bool(flagsMap[K8SSkipTlsVerify], false, "Disable TLS certificate verification for the Kubernetes API server")
	rootCmd.PersistentFlags().String(flagsMap[K8SCertificateAuthorityPath], "", "Path to the Kubernetes API server certificate authority PEM encoded file")

	rootCmd.PersistentFlags().String(flagsMap[KubeconfigPath], "oidc.kubeconfig", "Path to the generated kubeconfig file upon resulting login procedure to access the Kubernetes cluster")
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
