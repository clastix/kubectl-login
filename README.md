# Kubernetes Login Manager CLI

This `kubectl-login` is an utility to securely login Kubernetes clusters across multiple operating environments, including loacal setups and cloud providers, i.e. EKS, AKS, GKE. It can be used as `kubectl` plugin or as standalone binary.

Based on the configured authentication mechanism (e.g. TLS client, OIDC), it will login users in the Kubernetes clusters they are allowed to access and generate a `kubeconfig` for a chosen cluster.

## Features

- [ ] Authenticate with TLS client certificates
- [x] Authenticate against OIDC Server
    - [ ] Authorization Code Grant
    - [x] Authorization Code Grant with PKCE
    - [ ] Authorization with Resource Owner Password
    - [ ] Authorization with Credentials
    - [ ] Device Authorization Grant
- [ ] Authenticate against GKE
- [ ] Authenticate against EKS
- [ ] Authenticate against AKS
- [x] Create `kubeconfig`
- [x] Configure login parameters
- [ ] Store historical login parameters


## Installation

Download the release from the GitHub Release section according to your OS and architecture:

- [x] Darwin_x86_64
- [x] Linux_arm64
- [x] Linux_i386
- [x] Linux_x86_64
- [ ] Windows_x86_64

On Linux/MacOS systems, copy the binary somewhere on your `PATH`, and set it executable:

```bash
$ chmod u+x kubectl-login`
```

## Usage
Once you have installed `kubectl-login` you can see a list of the commands available by running:

```
$ kubectl login -h
kubectl-login is a CLI utility to discover and securely login Kubernetes clusters across multiple operating
environments, including local setups and cloud providers, i.e. EKS, AKS, GKE.

Based on the configured authentication mechanism (e.g. TLS client, OIDC), it will login users in the Kubernetes clusters
they are allowed to access and generate a kubeconfig for a chosen cluster.

Usage:
  login [flags]
  login [command]

Available Commands:
  get-token   Return a credential execution required by kubectl with the updated ID token
  help        Help about any command

Flags:
      --config string                   config file (default is $HOME/.kubectl-login.yaml)
  -h, --help                            help for login
      --k8s-api-server string           Endpoint of the Kubernetes API server to connect to
      --k8s-insecure-skip-tls-verify    Disable TLS certificate verification for the Kubernetes API server
      --k8s-server-ca-path string       Path to the Kubernetes API server certificate authority PEM encoded file
      --kubeconfig-path string          Path to the generated kubeconfig file upon resulting login procedure to access the Kubernetes cluster (default "oidc.kubeconfig")
      --oidc-client-id string           The OIDC client ID provided
      --oidc-insecure-skip-tls-verify   Disable TLS certificate verification for the OIDC server
      --oidc-server string              The OIDC server URL to connect to
      --oidc-server-ca-path string      Path to the OIDC server certificate authority PEM encoded file
  -v, --verbose                         Toggle the verbose logging

Use "login [command] --help" for more information about a command.
```

Create an initial setup:

```
$ kubectl login --k8s-api-server=https://kube-apiserver:6443 --k8s-server-ca-path=/path/to/k8s/ca.pem --oidc-server=https://sso.clastix.io --oidc-client-id=kubectl -v
2021-01-27T18:15:16.988Z        INFO    cmd/root.go:102 Starting the login procedure
2021-01-27T18:15:16.988Z        INFO    actions/oidc_config.go:63       Starting OIDC login with PKCE
2021-01-27T18:15:16.988Z        INFO    actions/oidc_config.go:74       Getting OIDC configuration from the server      {"OIDCServer": "https://sso.clastix.io"}
2021-01-27T18:15:17.022Z        INFO    actions/code_verifier.go:38     Generating PKCE Code Verifier and Challenge
2021-01-27T18:15:17.022Z        INFO    actions/code_verifier.go:39     PKCE code verifier generated    {"code": "PZD4n80AepGINMw1au4fMj73K0R38EXyPGd0QhmsIF3a3KRU3NBh2QwzSd9PAQ5dt1JifcbaixysCXIAQKhkV0lPituFgtTeWBIcWFmrfMCwvt8Cni2OP6vTc3sWOgPe"}
2021-01-27T18:15:17.023Z        INFO    actions/create_auth_uri.go:45   Creating authorization URI

Proceed to login to the following link using your browser:

https://sso.clastix.io/openid-connect/auth?access_type=offline&client_id=kubectl&code_challenge=EYpNK9lNI3g9ridirZLUxzZZC4uJPdIIdheVOYHZReY&code_challenge_method=S256&prompt=consent&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code&scope=openid+groups+offline_access&state=TDE5a90dfVLyeXxaHIbExowZoa344IztYcPXRgX0M

Type the verification code: *******************
2021-01-27T18:15:28.832Z        DEBUG   cmd/root.go:137 User input code is *******************
Your login procedure has been completed!

You can start interacting with your Kubernetes cluster using the generated kubeconfig file:
export KUBECONFIG=oidc.kubeconfig

Happy Kubernetes interaction!
```

The initial setup creates and stores configurations in the file `~/.kubectl-login.yaml`

```bash
kubernetes:
  ca:
    insecure: false
  endpoint: https://kube-apiserver:6443
  kubeconfig: oidc.kubeconfig
oidc:
  ca:
    insecure: false
  clientid: kubectl
  server: https://sso.clastix.io
token:
  endpoint: https://sso.clastix.io/openid-connect/token
  id: REDACTED
  refresh: REDACTED
```

A `kubeconfig` file is created as:

```yaml
apiVersion: v1
clusters:
  - cluster:
      server: https://kube-apiserver:6443
    name: kubernetes
contexts:
  - context:
      cluster: kubernetes
      user: oidc
    name: oidc
current-context: oidc
kind: Config
preferences: {}
users:
  - name: oidc
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
          - login
          - get-token
        command: kubectl
        env: null
        provideClusterInfo: false
...
```

To use it, export or copy in your default location

```
$ export KUBECONFIG=oidc.kubeconfig
$ kubectl --user=oidc get pods -n oil-production
NAME                       READY   STATUS    RESTARTS   AGE
example-5b64df8865-96f2p   1/1     Running   0          13h
example-5b64df8865-fg9mv   1/1     Running   0          13h
example-5b64df8865-z6ts9   1/1     Running   0          13h
```

You can start the login process any time by simply running:

```
$ kubectl login
```

## Contributions
`kubectl-login` is released with Apache 2 open source license. Contributions are very welcome!
