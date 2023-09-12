#!/bin/bash

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

$CURL --silent "https://us-docker.pkg.dev/v2/jarvice/images/jarvice-dragen-service/tags/list" | $JQ -r '.tags[]'
