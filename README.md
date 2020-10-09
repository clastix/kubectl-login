# Kubernetes Login Manager CLI

`kubectl-login` is a CLI utility to discover and securely login Kubernetes clusters across multiple operating environments, including loacal setups and cloud providers, i.e. EKS, AKS, GKE. It can be used as `kubectl` plugin or as standalone binary.

Based on the configured authentication mechanism, e.g. TLS client, OIDC, it will login users in the Kubernetes clusters they are allowed to access and generate a `kubeconfig` for a chosen cluster.

## Features

- [ ] Authenticate using TLS client certificates
- [x] Authenticate using OIDC against different IdP
- [ ] Authenticate against GKE
- [ ] Authenticate against EKS
- [ ] Authenticate against AKS
- [x] Generate a `kubeconfig` for a cluster
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

```bash
$ kubectl login -h
```

You can start the login process against your current Kubernetes cluster by running:

```bash
$ kubectl login
username: alice
password: *******

kubeconfig file is: alice-oidc.kubeconfig
to use it as alice export KUBECONFIG=alice-oidc.kubeconfig
Warning: the OIDC token will expire in 30 minutes!
```

TBD

## Contributions
`kubectl-login` is released with Apache2 open source license. Contributions are very welcome!