#!/bin/bash
VERSION=
DRAGEN_LIC=
PROJECT=""
LICENSE_NAME=""

function usage {
    cat <<EOF

Usage:
    $0 [options]

Options:
    --version               version for published containers
                            (required) 
    --dragen-lic-project    Google Marketplace license project
                            (required)
    --dragen-lic-name       Google Marketplace license name
                            (required)

Example:
    $0 --version 1.0

EOF
}

MAKE=$(type -p make)
if [ -z "$MAKE" ]; then
    cat <<EOF
Could not find 'make' in PATH. It may not be installed.
EOF
    exit 1
fi

DOCKER=$(type -p docker)
if [ -z "$DOCKER" ]; then
    cat <<EOF
Could not find 'docker' in PATH. It may not be installed.
EOF
    exit 1
fi

while [ $# -gt 0 ]; do
    case $1 in
        --help)
            usage
            exit 0
            ;;
        --version)
            VERSION=$2
            shift; shift
            ;;
        --dragen-lic-project)
            PROJECT=$2
            shift; shift
            ;;
        --dragen-lic-name)
            LICENSE_NAME=$2
            shift; shift
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

[ -z "$VERSION" ] && echo '--version required' && usage && exit 1
[ -z "$LICENSE_NAME" ] && echo '--dragen-lic-name required' && usage && exit 1
[ -z "$PROJECT" ] && echo '--dragen-lic-project required' && usage && exit 1

GCLOUD=$(type -p gcloud)
if [ -z "$GCLOUD" ]; then
    cat <<EOF
Could not find 'gcloud' in PATH. It may not be installed.
EOF
    exit 1
fi

JQ=$(type -p jq)
if [ -z "$JQ" ]; then
    cat <<EOF
Could not find 'jq' in PATH. It may not be installed.
EOF
    exit 1
fi

CURL=$(type -p curl)
if [ -z "$CURL" ]; then
    cat <<EOF
Could not find 'curl' in PATH. It may not be installed.
EOF
    exit 1
fi

GOOGLEAPI="https://www.googleapis.com/compute/v1"
LIC_PROJECT="/projects/$PROJECT/global/licenses"
DRAGEN_LIC=$($CURL -H "Authorization: Bearer $($GCLOUD auth print-access-token)" \
        "${GOOGLEAPI}${LIC_PROJECT}/${LICENSE_NAME}" 2>/dev/null \
        | $JQ -r .licenseCode)

[[ ! "$DRAGEN_LIC" =~ ^-?[0-9]+$ ]] && echo 'cannot find Google license ID' && exit 1

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "VERSION=$VERSION" > $SCRIPT_DIR/../VERSION.mk
echo "DRAGEN_LIC=$DRAGEN_LIC" >> $SCRIPT_DIR/../VERSION.mk

cd $SCRIPT_DIR/.. && $MAKE publish && rm VERSION.mk
