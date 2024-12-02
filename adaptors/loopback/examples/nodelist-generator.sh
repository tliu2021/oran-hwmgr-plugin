#!/bin/bash
#

PROG=$(basename "$0")
declare -A POOLS=()

USERNAME_BASE64=$(echo -n "admin" | base64)
PASSWORD_BASE64=$(echo -n "mypass" | base64)

function usage {
    cat <<EOF
Usage: ${PROG} ...
Parameters:
    --resourcepool <name:prefix:size>

Example:

${0} --resourcepool master:dummy-sp-64g:5 --resourcepool worker:dummy-dp-128g:3

EOF
    exit 1
}

function header {
    cat <<EOF
kind: ConfigMap
apiVersion: v1
metadata:
  name: loopback-adaptor-nodelist
  namespace: oran-hwmgr-plugin
data:
  resources: |
EOF
}

function resourcepools {
    echo "    resourcepools:"
    mapfile -t sorted_pools < <( IFS=$'\n'; sort -u <<<"${!POOLS[*]}" )
    for pool in "${sorted_pools[@]}"; do
        echo "      - ${pool}"
    done
}

function nodes {
    echo "    nodes:"
    group=0
    mapfile -t sorted_pools < <( IFS=$'\n'; sort -u <<<"${!POOLS[*]}" )
    for pool in "${sorted_pools[@]}"; do
        value="${POOLS[${pool}]}"
        prefix=$(echo "${value}" | awk -F: '{print $1}')
        size=$(echo "${value}" | awk -F: '{print $2}')
        group=$((group+1))

        for ((i=0;i<${size};i++)); do
            nodename=${prefix}-${i}
            ip="192.168.${group}.${i}"
            mac=$(printf "c6:b6:13:a0:%02x:%02x" ${group} ${i})

            cat <<EOF
      ${nodename}:
        poolID: ${pool}
        bmc:
          address: "idrac-virtualmedia+https://${ip}/redfish/v1/Systems/System.Embedded.1"
          username-base64: ${USERNAME_BASE64}
          password-base64: ${PASSWORD_BASE64}
        interfaces:
          - name: eth0
            label: bootable-interface
            macAddress: "${mac}"
EOF
        done
    done
}

#
# Process cmdline arguments
#

longopts=(
    "help"
    "resourcepool:"
)

longopts_str=$(IFS=,; echo "${longopts[*]}")

if ! OPTS=$(getopt -o "hp:" --long "${longopts_str}" --name "$0" -- "$@"); then
    usage
    exit 1
fi

eval set -- "${OPTS}"

while :; do
    case "$1" in
        -p|--resourcepool)
            value="$2"
            name=$(echo "${value}" | awk -F: '{print $1}')
            prefix=$(echo "${value}" | awk -F: '{print $2}')
            size=$(echo "${value}" | awk -F: '{print $3}')
            POOLS+=(["${name}"]="${prefix}:${size}")
            shift 2
            ;;
        --)
            shift
            break                                                                                                                                                                              ;;
        -h|--help)
            usage
            ;;
        *)
            usage
            ;;
    esac
done

if [ "${#POOLS[@]}" -eq 0 ]; then
    usage
fi

header
resourcepools
nodes

