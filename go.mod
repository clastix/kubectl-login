module github.com/clastix/kubectl-login

go 1.14

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/manifoldco/promptui v0.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)