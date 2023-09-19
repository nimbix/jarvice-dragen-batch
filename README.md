# Illumina DRAGEN batch engine

This project enables Google Batch processing of Illumina DRAGEN pipelines on Eviden powered FPGA infrastructure.

# Example

1. Upload sample files to Google Cloud Storage

*Note: Illumina sample files are ~90GB*

```bash
# Google Cloud Storage bucket for testing
BUCKET_NAME=
# Google Cloud project
PROJECT=
# get sample files
mkdir batch-test && cd batch-test
wget https://ilmn-dragen-giab-samples.s3.amazonaws.com/WGS/precisionFDA_v2_HG002/HG002.novaseq.pcr-free.35x.R1.fastq.gz
wget https://ilmn-dragen-giab-samples.s3.amazonaws.com/WGS/precisionFDA_v2_HG002/HG002.novaseq.pcr-free.35x.R2.fastq.gz
mkdir v8 && cd v8
wget https://ilmn-dragen-giab-samples.s3.amazonaws.com/Hashtable/hg38_altaware-cnv-graph-anchored.v8.tar
tar -xf hg38_altaware-cnv-graph-anchored.v8.tar && rm hg38_altaware-cnv-graph-anchored.v8.tar && cd ..
# create storage bucket for test
gcloud storage buckets --project $PROJECT create gs://$BUCKET_NAME
gcloud storage cp --project $PROJECT HG002.novaseq.pcr-free.35x.R1.fastq.gz gs://$BUCKET_NAME
gcloud storage cp --project $PROJECT HG002.novaseq.pcr-free.35x.R2.fastq.gz gs://$BUCKET_NAME
gcloud storage cp -r --project $PROJECT v8 gs://$BUCKET_NAME
```

2. [Create HMAC keys](https://cloud.google.com/storage/docs/authentication/managing-hmackeys)

3. Create Google Cloud Secrets

```bash
# JARVICE credentials provided during onboarding
JARVICE_API_USERNAME=
JARVICE_API_APIKEY=
# HMAC keys created in Step 2
S3_ACCESS_KEY=
S3_SECRET_KEY=
# Illumina license string
ILLUMINA_LIC_SERVER=
# Google cloud project
PROJECT=
# Google Cloud zone used for testing (e.g. us-central1)
ZONE=

printf "$JARVICE_API_USERNAME" | gcloud secrets create --project $PROJECT "jarviceApiUsername" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$JARVICE_API_APIKEY" | gcloud secrets create --project $PROJECT "jarviceApiKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$S3_ACCESS_KEY" | gcloud secrets create --project $PROJECT "batchS3AccessKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$S3_SECRET_KEY" | gcloud secrets create --project $PROJECT "batchS3SecretKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$ILLUMINA_LIC_SERVER" | gcloud secrets create --project $PROJECT "illuminaLicServer" --data-file=- --replication-policy=user-managed --locations=$ZONE
```

4. Prepare batch example
```bash
git clone https://github.com/nimbix/jarvice-dragen-batch
cd jarvice-dragen-batch/examples
# Google Cloud Storage bucket for testing
BUCKET_NAME=""
# Google Batch (https://cloud.google.com/batch) jobname
NAME="sample-batch-job"
# Google Cloud project
PROJECT=""
# Google Cloud zone (e.g. us-central1)
ZONE=""
# list available versions with: ../tools/list-versions.sh
VERSION=""
# DRAGEN application (e.g "illumina-dragen_3_7_8n")
JARVICE_DRAGEN_APP="illumina-dragen_3_7_8n"
cat > env.sh <<EOF
NAME="$NAME"
# Google Cloud project
PROJECT="$PROJECT"
# Google Cloud zone (e.g. us-central1)
ZONE="$ZONE"
# Google Cloud service account
SERVICE_ACCOUNT="default"
# list available versions with: ../tools/list-versions.sh
VERSION="$VERSION"
# DRAGEN application (e.g "illumina-dragen_3_7_8n")
JARVICE_DRAGEN_APP="$JARVICE_DRAGEN_APP"
JARVICE_API_USERNAME_SECRET="projects/${PROJECT}/secrets/jarviceApiUsername/versions/1"
JARVICE_API_APIKEY_SECRET="projects/${PROJECT}/secrets/jarviceApiKey/versions/1"
S3_ACCESS_KEY_SECRET="projects/${PROJECT}/secrets/batchS3AccessKey/versions/1"
S3_SECRET_KEY_SECRET="projects/${PROJECT}/secrets/batchS3SecretKey/versions/1"
ILLUMINA_LIC_SERVER_SECRET="projects/${PROJECT}/secrets/illuminaLicServer/versions/1"

DRAGEN_ARGS=(
-f
--enable-variant-caller true
--vc-emit-ref-confidence GVCF
--vc-enable-vcf-output true
--enable-duplicate-marking true
--enable-map-align true
--enable-map-align-output true
--vc-frd-max-effective-depth 40
--vc-enable-joint-detection true
--ht-alt-aware-validate false
--read-trimmers polyg
--soft-read-trimmers none
--logging-to-output-dir true
-1 s3://$BUCKET_NAME/HG002.novaseq.pcr-free.35x.R1.fastq.gz
-2 s3://$BUCKET_NAME/HG002.novaseq.pcr-free.35x.R2.fastq.gz
--RGID HG002
--RGSM HG002
-r s3://$BUCKET_NAME/v8
--output-file-prefix HG002_pure
--output-format CRAM
--output-directory s3://$BUCKET_NAME/output
)
EOF
```
5. Run example
```bash
./google-batch.sh
```
