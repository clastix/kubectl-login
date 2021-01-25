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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	yaml2 "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/tools/clientcmd/api/v1"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure your kubeconfig file to access the selected Kubernetes cluster",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = authCmd.RunE(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		config := &v1.Config{
			Kind:           "Config",
			APIVersion:     "v1",
			CurrentContext: "oidc",
			Clusters: []v1.NamedCluster{
				{
					Name: "kubernetes",
					Cluster: v1.Cluster{
						Server:                viper.GetString(ApiServer),
						InsecureSkipTLSVerify: true,
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
							Args:       []string{rootCmd.Use, tokenCmd.Use},
							APIVersion: "client.authentication.k8s.io/v1beta1",
						},
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		e := json.NewSerializerWithOptions(yaml2.SimpleMetaFactory{}, scheme, scheme, json.SerializerOptions{
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

		fmt.Println("")
		fmt.Println("Your login procedure has been completed and you can start interacting with your Kuberneter cluster!")
		fmt.Println("")
		fmt.Println("The generated kubeconfig file has been saved and can use it as following:")
		fmt.Println(fmt.Sprintf("export KUBECONFIG=%s", path))
		fmt.Println("")
		fmt.Println("Happy Kubernetes interaction!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().String("kubeconfig-path", "oidc.kubeconfig", "Path to save the generated kubeconfig file")
}
