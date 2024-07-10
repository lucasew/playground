export ZITI_HOME=$(pwd)/state
export ZITI_NETWORK=up-and-running

export ZITI_HOME=${ZITI_HOME}
export ZITI_CTRL_ADVERTISED_ADDRESS=127.0.0.1
export ZITI_CTRL_EDGE_ADVERTISED_HOST_PORT=127.0.0.1:1280
export ZITI_EDGE_CTRL_ADVERTISED_HOST_PORT=127.0.0.1:1280
export ZITI_PKI_CTRL_CERT=$ZITI_HOME/etc/ca/intermediate/certs/ctrl-client.cert.pem
export ZITI_PKI_CTRL_SERVER_CERT=$ZITI_HOME/etc/ca/intermediate/certs/ctrl-server.cert.pem
export ZITI_PKI_CTRL_KEY=$ZITI_HOME/etc/ca/intermediate/private/ctrl.key.pem
export ZITI_PKI_CTRL_CA=$ZITI_HOME/etc/ca/intermediate/certs/ca-chain.cert.pem
export ZITI_PKI_SIGNER_CERT=$ZITI_HOME/etc/ca/intermediate/certs/intermediate.cert.pem
export ZITI_PKI_SIGNER_KEY=$ZITI_HOME/etc/ca/intermediate/private/intermediate.key.decrypted.pem

export ADMIN_NAME=admin
export ADMIN_PW=admin

mkdir -p $ZITI_HOME/{db,etc/ca/intermediate/{certs,private}}

function ziti_setup {
  ziti pki create ca --pki-root=$ZITI_PKI --ca-file=$ZITI_CA_NAME
  ziti pki create server \
    --pki-root=$ZITI_PKI \
    --ca-name $ZITI_CA_NAME \
    --server-file "${ZITI_NETWORK}-ctrl-server" \
    --dns "${ZITI_CTRL_HOSTNAME}" \
    --ip "127.0.0.1" \
    --server-name "${ZITI_NETWORK} Controller"

  ziti create config controller --output $ZITI_HOME/db/ctrl-config.yml
  ziti controller edge init $ZITI_HOME/db/ctrl-config.yml -u $ADMIN_NAME -p $ADMIN_PW

  ziti edge create edge-router router01 --jwt-output-file $ZITI_HOME/router01.jwt --tunneler-enabled
  ziti create config router edge --routerName router01 --output $ZITI_HOME/db/router01-config.yml
}

function ziti_start_controller {
  ziti controller run $ZITI_HOME/db/ctrl-config.yml
}

function ziti_login_controller {
  ziti edge login -u $ADMIN_NAME -p $ADMIN_PW
}

function ziti_start_router {
  ziti router enroll --jwt $ZITI_HOME/router01.jwt $ZITI_HOME/db/router01-config.yml
}
