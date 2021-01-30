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

package oidc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/afero"
)

type CAReadFileError struct {
	error error
}

func (r CAReadFileError) Error() string {
	return fmt.Sprintf("Cannot read OIDC Certificate Authority file: %s", r.error.Error())
}

func NewCAReadFileError(error error) error {
	return &CAReadFileError{error: error}
}

type CAPoolError struct {
}

func (CAPoolError) Error() string {
	return "Cannot create CA Pool from PEM"
}

func NewOIDCCAPoolError() error {
	return &CAPoolError{}
}

type HTTPClient struct {
	http.Client
}

func NewHTTPClient(certificateAuthorityPath string, timeout time.Duration, insecureSkipVerify bool) (client *HTTPClient, err error) {
	return &HTTPClient{
		Client: http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					//nolint:gosec
					InsecureSkipVerify: insecureSkipVerify,
					RootCAs: func() (pool *x509.CertPool) {
						if len(certificateAuthorityPath) == 0 {
							return nil
						}
						b, e := afero.ReadFile(afero.NewOsFs(), certificateAuthorityPath)
						if e != nil {
							err = NewCAReadFileError(e)
							return nil
						}
						pool = x509.NewCertPool()
						if !pool.AppendCertsFromPEM(b) {
							err = NewOIDCCAPoolError()
							return nil
						}
						return
					}(),
				},
			},
		},
	}, err
}
