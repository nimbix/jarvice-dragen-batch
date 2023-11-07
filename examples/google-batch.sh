#!/bin/bash

JARVICE_API_URL="https://illumina.nimbix.net/api"
JARVICE_MACHINE_TYPE=nx1
# Google Batch (https://cloud.google.com/batch) jobname
NAME=
# Google Cloud project
PROJECT=
# Google Cloud zone (e.g. "us-central1")
ZONE=
# Google Cloud service account
SERVICE_ACCOUNT=
# list available versions with: ../tools/list-versions.sh
VERSION=
# DRAGEN application (e.g "illumina-dragen_3_7_8n")
JARVICE_DRAGEN_APP=
# job priority (normal, high, or highest)
JARVICE_JOB_PRIORITY="normal"
# secrets maintained by Google Secret Manager (https://cloud.google.com/secret-manager)
# format: projects/${PROJECT}/secrets/<secret-name>/versions/1
JARVICE_API_USERNAME_SECRET=
JARVICE_API_APIKEY_SECRET=
S3_ACCESS_KEY_SECRET=
S3_SECRET_KEY_SECRET=
ILLUMINA_LIC_SERVER_SECRET=

DRAGEN_ARGS=(
# one dragen argument per line
# -f
# --logging-to-output-dir true
# ...
)
# optional environment file
source env.sh

# no edits required past this line

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

suffix=$(echo $RANDOM | md5sum | head -c 8)

JOBNAME="$NAME-$suffix"

batch_json=$(cat <<EOD
{
  "name": "projects/$PROJECT/locations/$ZONE/jobs/$JOBNAME",
  "taskGroups": [
    {
      "taskCount": "1",
      "parallelism": "1",
      "taskSpec": {
        "computeResource": {
          "cpuMilli": "1000",
          "memoryMib": "512"
        },
        "runnables": [
          {
            "environment": {
              "secretVariables": {
                "JARVICE_API_USER": "$JARVICE_API_USERNAME_SECRET",
                "JARVICE_API_KEY": "$JARVICE_API_APIKEY_SECRET",
                "S3_ACCESS_KEY": "$S3_ACCESS_KEY_SECRET",
                "S3_SECRET_KEY": "$S3_SECRET_KEY_SECRET",
                "ILLUMINA_LIC_SERVER": "$ILLUMINA_LIC_SERVER_SECRET"
              }
            },
            "container": {
              "imageUri": "us-docker.pkg.dev/jarvice/images/jarvice-dragen-service:$VERSION",
              "entrypoint": "/usr/local/bin/entrypoint",
              "commands": [
                "--api-host", "$JARVICE_API_URL",
                "--machine", "$JARVICE_MACHINE_TYPE",
                "--dragen-app", "$JARVICE_DRAGEN_APP",
                "--google-sa", "$SERVICE_ACCOUNT",
                "--job-priority", "$JARVICE_JOB_PRIORITY",
                "--"
              ],
              "volumes": []
            }
          }
        ],
        "volumes": []
      }
    }
  ],
  "allocationPolicy": {
    "instances": [
      {
        "policy": {
          "provisioningModel": "STANDARD",
          "machineType": "e2-micro"
        }
      }
    ]
  },
  "logsPolicy": {
    "destination": "CLOUD_LOGGING"
  }
}
EOD
)

for str in ${DRAGEN_ARGS[@]}; do
  batch_json=$(echo $batch_json | $JQ --arg arg "$str" '.taskGroups[0].taskSpec.runnables[0].container.commands[.taskGroups[0].taskSpec.runnables[0].container.commands | length] |= . + $arg');
done

echo $batch_json | $GCLOUD beta batch jobs submit  --project $PROJECT $JOBNAME --location $ZONE --config -
