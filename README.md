# Illumina DRAGEN batch engine

This project enables Google Batch processing of Illumina DRAGEN pipelines on Eviden powered FPGA infrastructure.

There are multiple ways to submit Illumina DRAGEN pipelines:

 - Script Submission
 - [Google Cloud Batch submission](batch-submission/README.md)
 - [Cromwell submission](cromwell-submission/README.md)
 - [Nextflow submission](nextflow-submission/README.md)

# Script Submission Example

1. Upload sample files to Google Cloud Storage

*Note: Illumina sample files are ~90GB*

```bash
#!/bin/bash

# Google Cloud Storage bucket for testing
BUCKET_NAME=<GCS bucket name>
# Google Cloud project
PROJECT=<GCP Project name>

# create storage bucket for test
gcloud storage buckets --project $PROJECT create gs://$BUCKET_NAME

#method 1 - use storage transfer service

# There is already a public sample data bucket in GCS with reference from Illumina. 
# IAM - give the project-<project-id>@@storage-transfer-service.iam.gserviceaccount.com role: Storage Legacy Bucket Writer
gcloud transfer jobs create gs://thomashk-public-illumina-sample gs://$BUCKET_NAME

# method 2 - copy data from GCP sample bucket to the new bucket manaually:

# get sample fastq files
wget https://storage.googleapis.com/thomashk-public-illumina-sample/HG002.novaseq.pcr-free.35x.R1.fastq.gz
wget https://storage.googleapis.com/thomashk-public-illumina-sample/HG002.novaseq.pcr-free.35x.R2.fastq.gz 

# Downloading Illumina DRAGEN Multigenome Graph Reference - hg38
# https://support.illumina.com/downloads/dragen-reference-genomes-hg38.html

# Reference for DRAGEN v4.2 
#mkdir 4_2_reference && cd 4_2_reference
#wget https://webdata.illumina.com/downloads/software/dragen/references/genome-files/hg38-alt_masked.cnv.graph.hla.rna-9-r3.0-1.tar.gz
#gunzip hg38-alt_masked.cnv.graph.hla.rna-9-r3.0-1.tar.gz
#tar -xvf hg38-alt_masked.cnv.graph.hla.rna-9-r3.0-1.tar
#cd ..
# Reference for DRAGEN v4.0
#mkdir 4_0_reference && cd 4_0_reference
#wget https://webdata.illumina.com/downloads/software/dragen/hg38%2Balt_masked%2Bcnv%2Bgraph%2Bhla%2Brna-8-r2.0-1.run
#./hg38%2Balt_masked%2Bcnv%2Bgraph%2Bhla%2Brna-8-r2.0-1.run
#cd ..
# Reference for DRAGEN v3.10
#mkdir 3_10_reference && cd 3_10_reference
#wget https://webdata.illumina.com/downloads/software/dragen/hg38_alt_masked_graph_v2%2Bcnv%2Bgraph%2Brna-8-1644018559-1.run
#./hg38_alt_masked_graph_v2%2Bcnv%2Bgraph%2Brna-8-1644018559-1.run
#cd ..
# Reference for DRAGEN v3.7 or older
#mkdir 3_7_reference && cd 3_7_reference
#wget https://s3.amazonaws.com/webdata.illumina.com/downloads/software/dragen/references/genome-files/hg38/hg38_alt_aware%2Bcnv%2Bgraph%2Brna-8-r1.0-0.run
#./hg38_alt_aware%2Bcnv%2Bgraph%2Brna-8-r1.0-0.run
#cd ..

# Downloading the data from local to the GCS bucket:
gcloud storage cp --project $PROJECT HG002.novaseq.pcr-free.35x.R1.fastq.gz gs://$BUCKET_NAME
gcloud storage cp --project $PROJECT HG002.novaseq.pcr-free.35x.R2.fastq.gz gs://$BUCKET_NAME
# select the reference needed and uncomment the line below:
#gcloud storage cp -r --project $PROJECT 4_2_reference gs://$BUCKET_NAME
#gcloud storage cp -r --project $PROJECT 4_0_reference gs://$BUCKET_NAME
#gcloud storage cp -r --project $PROJECT 3_10_reference gs://$BUCKET_NAME
#gcloud storage cp -r --project $PROJECT 3_9_reference gs://$BUCKET_NAME
#gcloud storage cp -r --project $PROJECT 3_7_reference gs://$BUCKET_NAME
```

2. [Create HMAC keys](https://cloud.google.com/storage/docs/authentication/managing-hmackeys)

3. Create Google Cloud Secrets

```bash
#!/bin/bash

# JARVICE credentials provided during onboarding
JARVICE_API_USERNAME=
JARVICE_API_APIKEY=
# HMAC keys created in Step 2
S3_ACCESS_KEY=
S3_SECRET_KEY=
# Illumina license string
ILLUMINA_LIC_SERVER=
# Google cloud project
PROJECT=<GCP Project name>
# Google Cloud zone used for testing (e.g. us-central1)
ZONE=us-central1

printf "$JARVICE_API_USERNAME" | gcloud secrets create --project $PROJECT "jarviceApiUsername" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$JARVICE_API_APIKEY" | gcloud secrets create --project $PROJECT "jarviceApiKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$S3_ACCESS_KEY" | gcloud secrets create --project $PROJECT "batchS3AccessKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$S3_SECRET_KEY" | gcloud secrets create --project $PROJECT "batchS3SecretKey" --data-file=- --replication-policy=user-managed --locations=$ZONE
printf "$ILLUMINA_LIC_SERVER" | gcloud secrets create --project $PROJECT "illuminaLicServer" --data-file=- --replication-policy=user-managed --locations=$ZONE
```

4. Prepare batch example file - env.sh
```bash
# This is a sample env.sh file. Please update all the GCP project name and bucket name before using.
NAME="sample-batch-job"
# Google Cloud project
PROJECT="<GCP Project name>"
# Google Cloud zone (e.g. us-central1)
ZONE="us-central1"
# Google Cloud service account
SERVICE_ACCOUNT="<project id>-compute@developer.gserviceaccount.com"
# list available versions with: ../tools/list-versions.sh
VERSION="1.0-rc.5"
# DRAGEN application (e.g "illumina-dragen_3_7_8n")
JARVICE_DRAGEN_APP="illumina-dragen_4_2_4n"
JARVICE_API_USERNAME_SECRET="projects/<GCP Project name>/secrets/jarviceApiUsername/versions/latest"
JARVICE_API_APIKEY_SECRET="projects/<GCP Project name>/secrets/jarviceApiKey/versions/latest"
S3_ACCESS_KEY_SECRET="projects/<GCP Project name>/secrets/batchS3AccessKey/versions/latest"
S3_SECRET_KEY_SECRET="projects/<GCP Project name>/secrets/batchS3SecretKey/versions/latest"
ILLUMINA_LIC_SERVER_SECRET="projects/<GCP Project name>/secrets/illuminaLicServer/versions/latest"

DRAGEN_ARGS=(
-f
-r s3://<GCS Bucket name>/4_2_reference
-1 s3://<GCS Bucket name>/HG002.novaseq.pcr-free.35x.R1.fastq.gz
-2 s3://<GCS Bucket name>/HG002.novaseq.pcr-free.35x.R2.fastq.gz
--RGID HG002 
--RGSM HG002 
--output-directory s3://<GCP Bucket name>/output2
--output-file-prefix HG002_4_2
--enable-map-align true
--enable-map-align-output true
--output-format CRAM
--enable-duplicate-marking true
--enable-variant-caller true
--vc-enable-vcf-output true
--vc-emit-ref-confidence GVCF
--vc-frd-max-effective-depth 40
--vc-enable-joint-detection true
--read-trimmers polyg
--soft-read-trimmers none
)
```
5. Run example
```bash
./google-batch.sh
```
