# Kubernetes Login Manager CLI

`kubectl-login` is a CLI utility to discover and securely login Kubernetes clusters across multiple operating environments, including loacal setups and cloud providers, i.e. EKS, AKS, GKE. It can be used as `kubectl` plugin or as standalone binary.

Based on the configured authentication mechanism, e.g. TLS client, OIDC, it will login users in the Kubernetes clusters they are allowed to access and generate a `kubeconfig` for a chosen cluster.

## Features

- [ ] Authenticate using TLS client certificates
- [x] Authenticate using OIDC against different IdP
    - [x] Authorization Code Grant
    - [x] Authorization Code Grant with PKCE
    - [x] Authorization with Resource Owner Password
- [ ] Authenticate against GKE
- [ ] Authenticate against EKS
- [ ] Authenticate against AKS
- [x] Update current `kubeconfig`
- [ ] Configure login parameters
- [ ] Store historical login parameters


## Installation
Install [curl](https://github.com/curl/curl) and [jq](https://stedolan.github.io/jq/) dependencies if you don't already in your system.

Copy the `kubectl-login` script somewhere on your `PATH`, and set it executable:

```bash
$ chmod u+x kubectl-login`
```

Make sure to create a configuration file in your `~/.kube/oidc.conf` directory

```bash
OIDC_SERVER=https://sso.clastix.io/auth/realms/caas
OIDC_CLIENT_ID=kubectl
PKCE=enabled
CODE_CHALLENGE_METHOD=S256
REDIRECT_URI=urn:ietf:wg:oauth:2.0:oob
```


## Usage
Once you have installed `kubectl-login` you can see a list of the commands available by running:

```bash
$ kubectl login -h
Usage: /usr/local/bin/kubectl-login [OPTIONS]

    OPTIONS:
    --help     display usage
    --login    login to the OIDC Server
    --token    return a valid id_token 
    --setup    configure current kubeconfig 
 
```

Create an initial setup:

```bash
$ kubectl oidc --setup
[Thu Jan  7 13:04:55 CET 2021][INFO] Checking if prerequisites are installed
[Thu Jan  7 13:04:55 CET 2021][INFO] Starting OIDC login with PKCE
[Thu Jan  7 13:04:55 CET 2021][INFO] Getting OIDC configuration from https://sso.clastix.io/auth/realms/caas
[Thu Jan  7 13:05:01 CET 2021][INFO] Generating PKCE Code Verifier and Challenge
[Thu Jan  7 13:05:02 CET 2021][INFO] Creating authorization URI

Go to the following link in your browser:

https://sso.clastix.io/auth/realms/caas/protocol/openid-connect/auth?querystring

Enter verification code: **************

[Thu Jan  7 13:05:11 CET 2021][INFO] Requesting token from https://sso.clastix.io/auth/realms/caas
[Thu Jan  7 13:05:12 CET 2021][INFO] Saving token to cache
[Thu Jan  7 13:05:12 CET 2021][INFO] Saving configuration to /Users/adriano/.kube/oidc.conf
[Thu Jan  7 13:05:12 CET 2021][INFO] Configuring current kubeconfig
User "oidc" set.

Make sure you can access the Kubernetes cluster:

      $ kubectl --user=oidc get pods

You can switch the current context:

      $ kubectl config set-context --current --user=oidc

Or you can set a new context:

      $ kubectl config set-context oidc --user=oidc --cluster=nickname
      $ kubectl config use-context oidc

```

And use it:

```bash
$ kubectl --user=oidc get pods -n oil-production
NAME                       READY   STATUS    RESTARTS   AGE
example-5b64df8865-96f2p   1/1     Running   0          13h
example-5b64df8865-fg9mv   1/1     Running   0          13h
example-5b64df8865-z6ts9   1/1     Running   0          13h
```

You can start the login process any time by simply running:

```bash
$ kubectl login --login
```

Check your current `kubeconfig` file

```yaml
kind: Config
preferences: {}
users:
- name: oidc
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc
      - --token
      command: kubectl
      env: null
...
```

## Contributions
`kubectl-login` is released with Apache2 open source license. Contributions are very welcome!