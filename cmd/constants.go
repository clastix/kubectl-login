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

const (
	// Kubernetes viper keys
	K8SAPIServer                = "kubernetes.endpoint"
	KubeconfigPath              = "kubernetes.kubeconfig"
	K8SSkipTLSVerify            = "kubernetes.ca.insecure"
	K8SCertificateAuthorityPath = "kubernetes.ca.path"
	// OIDC viper keys
	OIDCServer               = "oidc.server"
	OIDCClientID             = "oidc.clientid"
	OIDCSkipTLSVerify        = "oidc.ca.insecure"
	OIDCCertificateAuthority = "oidc.ca.path"
	// Token viper keys
	TokenID       = "token.id"
	TokenRefresh  = "token.refresh"
	TokenEndpoint = "token.endpoint"
)

var (
	flagsMap = map[string]string{
		// OIDC flags
		OIDCServer:               "oidc-server",
		OIDCClientID:             "oidc-client-id",
		OIDCSkipTLSVerify:        "oidc-insecure-skip-tls-verify",
		OIDCCertificateAuthority: "oidc-server-ca-path",
		// Kubernetes flags
		K8SAPIServer:                "k8s-api-server",
		K8SSkipTLSVerify:            "k8s-insecure-skip-tls-verify",
		K8SCertificateAuthorityPath: "k8s-server-ca-path",
		KubeconfigPath:              "kubeconfig-path",
	}
)
