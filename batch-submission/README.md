# Launching the Illumina-DRAGEN job with native batch submission

You can submit Illumina DRAGEN pipelines with single JSON file into Google Batch. Provided you have Batch API enabled, GCP IAM permissions and Google Cloud Secrets setup completed.

It is simple to use single JSON file to config. 
- Update the json file provided with the correct GCP project and secret information:`batchsubmit.json`

```
{
  "taskGroups": [
    {
      "taskSpec": {
        "runnables": [
              {
                "environment": {
                  "secretVariables": {
                    "JARVICE_API_USER": "projects/GCPPROJECTID/secrets/jarviceApiUsername/versions/latest",
                    "JARVICE_API_KEY": "projects/GCPPROJECTID/secrets/jarviceApiKey/versions/latest",
                    "S3_ACCESS_KEY": "projects/GCPPROJECTID/secrets/batchS3AccessKey/versions/latest",
                    "S3_SECRET_KEY": "projects/GCPPROJECTID/secrets/batchS3SecretKey/versions/latest",
                    "ILLUMINA_LIC_SERVER": "projects/GCPPROJECTID/secrets/illuminaLicServer/versions/latest"
              }
            },          
	   "container": {
              "imageUri": "us-docker.pkg.dev/jarvice/images/jarvice-dragen-service:1.0-rc.5",
              "entrypoint": "/usr/local/bin/entrypoint",
              "commands": [
		"--api-host", "https://illumina.nimbix.net/api",
		"--machine", "nx1",
		"--dragen-app", "illumina-dragen_4_2_4n",
		"--google-sa", "GCPPROJECTNUMBER-compute@developer.gserviceaccount.com",
		"--", "-f",
		"-1", "s3://jarvice-dragen-batch/HG002.novaseq.pcr-free.35x.R1.fastq.gz",
		"-2", "s3://jarvice-dragen-batch/HG002.novaseq.pcr-free.35x.R2.fastq.gz",
		"--RGID", "HG002",
		"--RGSM", "HG002",
		"-r", "s3://jarvice-dragen-batch/4_2_reference",
		"--enable-map-align", "true",
		"--enable-map-align-output", "true",
		"--enable-duplicate-marking", "true",
		"--output-format", "CRAM",
		"--enable-variant-caller", "true",
		"--vc-emit-ref-confidence"," GVCF",
		"--vc-frd-max-effective-depth", "40",
		"--vc-enable-joint-detection", "true",
		"--read-trimmers", "polyg",
		"--soft-read-trimmers", "none",
		"--vc-enable-vcf-output", "true",
		"--output-file-prefix", "HG002_4_2",
		"--output-directory", "s3://jarvice-dragen-batch/output2"
                ]
             }
    		}
	  ]
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
```

- Use the gcloud command to submit the DRAGEN job. The job name has to be unique in the GCP project

`gcloud batch jobs submit illuminatest-date-1 --location us-central1 --config batchsubmit.json`
