# Kubernetes Login Manager CLI

`kubectl-login` is a CLI utility to discover and securely login Kubernetes clusters across multiple operating environments, including loacal setups and cloud providers, i.e. EKS, AKS, GKE. It can be used as `kubectl` plugin or as standalone binary.

Based on the configured authentication mechanism, e.g. TLS client, OIDC, it will login users in the Kubernetes clusters they are allowed to access and generate a `kubeconfig` for a chosen cluster.

## Features

- [ ] Authenticate using TLS client certificates
- [x] Authenticate using OIDC
    - [x] Authorization Code Grant
    - [x] Authorization Code Grant with PKCE
    - [x] Authorization with Resource Owner Password
- [ ] Authenticate against GKE
- [ ] Authenticate against EKS
- [ ] Authenticate against AKS
- [x] Create `kubeconfig`
- [ ] Configure login parameters
- [ ] Store historical login parameters


## Installation
Install [curl](https://github.com/curl/curl) and [jq](https://stedolan.github.io/jq/) dependencies if you don't already in your system.

Copy the `kubectl-login` script somewhere on your `PATH`, and set it executable:

```bash
$ chmod u+x kubectl-login`
```

## Usage
Once you have installed `kubectl-login` you can see a list of the commands available by running:

```
$ kubectl login -h
Usage: /usr/local/bin/kubectl-login [OPTIONS]

    OPTIONS:
    --help     display usage
    --login    login to the OIDC Server
    --token    return a valid id_token 
    --setup    configure current kubeconfig 
 
```

Create an initial setup:

```
$ kubectl login --setup
[Tue Jan 12 18:26:21 CET 2021][INFO] Checking if prerequisites are installed
[Tue Jan 12 18:26:21 CET 2021][INFO] Starting OIDC login with PKCE
[Tue Jan 12 18:26:21 CET 2021][INFO] Setting up configuration
Kubernetes APIs Server: https://cmp.clastix.io
OIDC Server URL: https://sso.clastix.io/auth/realms/caas
OIDC Client ID: kubectl
[Tue Jan 12 18:26:50 CET 2021][INFO] Getting OIDC configuration from https://sso.clastix.io/auth/realms/caas
[Tue Jan 12 18:26:50 CET 2021][INFO] Generating PKCE Code Verifier and Challenge
[Tue Jan 12 18:26:50 CET 2021][INFO] Creating authorization URI

Go to the following link in your browser:

https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/auth?response_type=code&client_id=kubectl&redirect_uri=urn:ietf:wg:oauth:2.0:oob&scope=openid+groups+offline_access&state=LnA6VFW3L0NmnrxDUNg52gKwbM3yI5MnDtWxKtPV4&prompt=consent&code_challenge=1icW0jFFHTcOU5QBZO_nGKiFdeg8Uo5nuel3MbCdoSo&code_challenge_method=S256&access_type=offline

Enter verification code: **************

[Tue Jan 12 18:27:07 CET 2021][INFO] Requesting token from https://sso.clastix.io/auth/realms/caas
[Tue Jan 12 18:27:07 CET 2021][INFO] Saving token to cache
[Tue Jan 12 18:27:07 CET 2021][INFO] Saving configuration to ~/.kube/oidc.conf
[Tue Jan 12 18:27:07 CET 2021][INFO] Creating kubeconfig

Make sure you can access the Kubernetes cluster:

      $ export KUBECONFIG=oidc.kubeconfig
      $ kubectl --user=oidc get pods
```

The initial setup creates and stores configurations in the file `~/.kube/oidc.conf`

```bash
API_SERVER=https://cmp.clastix.io
OIDC_SERVER=https://sso.clastix.io/auth/realms/caas
OIDC_CLIENT_ID=kubectl
CODE_CHALLENGE_METHOD=S256
AUTH_ENDPOINT=https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/auth
TOKEN_ENDPOINT=https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/token
INTROSPECTION_ENDPOINT=https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/token/introspect
USERINFO_ENDPOINT=https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/userinfo
END_SESSION_ENDPOINT=https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/logout
```

A `kubeconfig` file is created as:

```yaml
kind: Config
preferences: {}
users:
- name: oidc
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - login
      - --token
      command: kubectl
      env: null
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
$ kubectl login --login
```



## Contributions
`kubectl-login` is released with Apache2 open source license. Contributions are very welcome!