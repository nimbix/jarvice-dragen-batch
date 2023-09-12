#!/bin/bash

VERSION=
DRAGEN_LIC=

function usage {
    cat <<EOF

Usage:
    $0 [options]

Options:
    --version               version for published containers
                            (required) 
    --dragen-lic            Google Marketplace license id (integer)
                            (required)

Example:
    $0 --version 1.0 --dragen-lic 12345678

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
        --dragen-lic)
            DRAGEN_LIC=$2
            shift; shift
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

[ -z "$VERSION" ] && echo '--version required' && usage && exit 1

[ -z "$DRAGEN_LIC" ] && echo '--dragen-lic required' && usage && exit 1

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "VERSION=$VERSION" > $SCRIPT_DIR/../VERSION.mk
echo "DRAGEN_LIC=$DRAGEN_LIC" >> $SCRIPT_DIR/../VERSION.mk

cd $SCRIPT_DIR/.. && $MAKE publish && rm VERSION.mk
