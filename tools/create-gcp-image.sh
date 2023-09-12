#!/bin/bash
set -e
IMAGE_NAME=
PROJECT=
LICENSE=
DESCRIPTION=
SRC_IMAGE="cos-stable-105-17412-156-30"
SRC_PROJECT="cos-cloud"

function usage {
    cat <<EOF

Usage:
    $0 [options]

Options:
    --image-name           GCP image name
                           (required)
    --project              GCP project
                           (required)
    --license              Google Marketplace License (string)
                           (required)
    --description          GCP image description
                           (required)
    --src-image            Source GCP image
                           (default: $SRC_IMAGE)
    --src-project          Source GCP project
                           (default: $SRC_PROJECT)

Example:
    $0 --image-name "dragen-vm-image" \
        --project "my-marketplace-project" \
        --license "my-marketplace-license-string" \
        --description "dragen marketplace image" 

NOTE:
    Google Marketplace images must be public !!!
    Do not include any sensative data

EOF
}

GCLOUD=$(type -p gcloud)
if [ -z "$GCLOUD" ]; then
	    cat <<EOF
Could not find 'gcloud' in PATH. It may not be installed.
EOF
    exit 1
fi

while [ $# -gt 0 ]; do
    case $1 in
    --help)
	    usage
	    exit 0
	    ;;
	--image-name)
	    IMAGE_NAME=$2
	    shift; shift
	    ;;
	--project)
	    PROJECT=$2
	    shift; shift
	    ;;
	--license)
	    LICENSE=$2
	    shift; shift
	    ;;
	--description)
	    DESCRIPTION=$2
	    shift; shift
	    ;;
    	--src-image)
	    SRC_IMAGE=$2
	    shift; shift
	    ;;
    	--src-project)
	    SRC_PROJECT=$2
	    shift; shift
	    ;;
	*)
	    usage
	    exit 1
	    ;;
    esac
done

[ -z "$IMAGE_NAME" ] && echo '--image-name required' && usage && exit 1

[ -z "$PROJECT" ] && echo '--project required' && usage && exit 1

[ -z "$LICENSE" ] && echo '--license required' && usage && exit 1

[ -z "$DESCRIPTION" ] && echo '--description required' && usage && exit 1

# create Google Marketplace image
$GCLOUD compute images create "$IMAGE_NAME" \
    --project "$PROJECT" \
    --source-image "projects/${SRC_PROJECT}/global/images/${SRC_IMAGE}" \
    --licenses "projects/${PROJECT}/global/licenses/${LICENSE}" \
    --description "$DESCRIPTION"
# NOTE: Google Marketplace images must be public
$GCLOUD compute images add-iam-policy-binding "$IMAGE_NAME" \
    --project "$PROJECT" \
    --member "allAuthenticatedUsers" \
    --role "roles/compute.imageUser"
