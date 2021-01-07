#!/bin/bash
USERNAME=$1
PASSWORD=$2
CLUSTER=$3
KEYCLOAK_URL=https://sso.clastix.io
KEYCLOAK_REALM=caas
KEYCLOAK_CLIENT_ID=kubernetes
KEYCLOAK_CLIENT_SECRET=*******

# Check if jq is installed
if [ ! -x "$(command -v jq)" ]; then
    echo "Error: jq not found"
    exit 1
fi


# Check if curl is installed
if [ ! -x "$(command -v curl)" ]; then
    echo "Error: curl not found - install it before to run" $0
    exit 1
fi

if [ "${CLUSTER}" = "" ];then
        read -p "cluster: " CLUSTER
fi

if [ "${USERNAME}" = "" ];then
        read -p "username: " USERNAME
fi

if [ "${PASSWORD}" = "" ];then
        read -sp "password: " PASSWORD
fi


KEYCLOAK_TOKEN_URL=${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token 

echo
echo "# Getting a token ..."

TOKEN=`curl -k -s ${KEYCLOAK_TOKEN_URL} \
  -d grant_type=password \
  -d response_type=id_token \
  -d scope="openid website" \
  -d client_id=${KEYCLOAK_CLIENT_ID} \
  -d client_secret=${KEYCLOAK_CLIENT_SECRET} \
  -d username=${USERNAME} \
  -d password=${PASSWORD}`

RET=$?
if [ "$RET" != "0" ];then
        echo "# Error ($RET) ==> ${TOKEN}";
        exit ${RET}
fi

ERROR=`echo ${TOKEN} | jq .error -r`
if [ "${ERROR}" != "null" ];then
        echo "# Failed ==> ${TOKEN}" >&2
        exit 1
fi

ID_TOKEN=`echo ${TOKEN} | jq .id_token -r`
ACCESS_TOKEN=`echo ${TOKEN} | jq .access_token -r`
REFRESH_TOKEN=`echo ${TOKEN} | jq .refresh_token -r`

echo
echo "got the access_token:"
curl -s -k --user ${KEYCLOAK_CLIENT_ID}:${KEYCLOAK_CLIENT_SECRET} \
    -d token=${ACCESS_TOKEN} ${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token/introspect | jq

echo
echo "got the id_token:"
curl -s -k --user ${KEYCLOAK_CLIENT_ID}:${KEYCLOAK_CLIENT_SECRET} \
    -d token=${ID_TOKEN} ${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token/introspect | jq
echo

echo "got the refresh_token:"
curl -s -k --user ${KEYCLOAK_CLIENT_ID}:${KEYCLOAK_CLIENT_SECRET} \
    -d token=${REFRESH_TOKEN} ${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token/introspect | jq
echo

echo "user info:"
curl -s -k -d access_token=${ACCESS_TOKEN} ${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/userinfo | jq

# Create the kubeconfig file
cat > ${USERNAME}-oidc.kubeconfig <<EOF
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://${CLUSTER}
  name: ${CLUSTER}
contexts:
- context:
    cluster: ${CLUSTER}
    user: ${USERNAME}
  name: ${USERNAME}-oidc@${CLUSTER}
current-context: ${USERNAME}-oidc@${CLUSTER}
kind: Config
preferences: {}
users:
- name: ${USERNAME}
  user:
    auth-provider:
      config:
        client-id: ${KEYCLOAK_CLIENT_ID}
        client-secret: ${KEYCLOAK_CLIENT_SECRET}
        extra-scopes: groups
        id-token: ${ID_TOKEN}
        idp-issuer-url: ${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}
        refresh-token: ${REFRESH_TOKEN}
        idp-certificate-authority: rootCA.crt
      name: oidc
EOF

echo "kubeconfig file is:" ${USERNAME}-oidc.kubeconfig
echo "to use it as" ${USERNAME} "export KUBECONFIG="${USERNAME}-oidc.kubeconfig
