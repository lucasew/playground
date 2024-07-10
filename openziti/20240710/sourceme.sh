export ZITI_NETWORK=up-and-running
export ZITI_HOME=$(pwd)/state

export ZITI_HOME=${ZITI_HOME}
export ZITI_NETWORK=${ZITI_NETWORK}
export ZITI_ID="${ZITI_HOME}/identities.yml"
export ZITI_CA_NAME="${ZITI_NETWORK}"
export ZITI_PKI="${ZITI_HOME}/pki"
export ZITI_CA_FILE="${ZITI_PKI}/${ZITI_CA_NAME}/certs/${ZITI_CA_NAME}.cert"
export ZITI_CTRL_HOSTNAME="${ZITI_NETWORK}-ctrl.ziti.netfoundry.io"
export ZITI_ER01_HOSTNAME="${ZITI_NETWORK}-er01.ziti.netfoundry.io"
export ZITI_R01_HOSTNAME="${ZITI_NETWORK}-r01.ziti.netfoundry.io"
export ZITI_EDGE_API_PORT=1280
export ZITI_EDGE_API_HOSTNAME="${ZITI_CTRL_HOSTNAME}:${ZITI_EDGE_API_PORT}"

mkdir -p $ZITI_HOME/db
mkdir -p $ZITI_PKI
