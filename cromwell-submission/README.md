# Submitting Illumina DRAGEN to Batch via Cromwell

## Steps

1. You still need to make prepare data and secrets mentioned in this git.
	Please check with the following:
	- Batch API enabled
	- GCP IAM permission for Batch job submission
	- Data uploaded to the Google Cloud Storage (GCS) Bucket
	- Google Cloud Secrets setup completed with JARVICE API Username, APIKEY, GCS S3_Access_key, GCS S3_Secret_key, Illumina License server URI

2. Prepare the single JSON file. 
	This is the single files to submit a new Illumina DRAGEN job into Batch.  This is an example we use in the batch-submission directory:
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
  
  After editing the file, you can upload this json file into your GCS bucket
  
3. Provided that you can launch cromwell job into GCP batch. This means you have the correct JAVA version and a cromwell configuration.
	Please follow the instructions for Cromwell installation with Batch backend.  
	We provided a sample configuration - gcpbatchillumina.conf

4. Prepare your wdl file:
The task has couple steps including
	- downloading the json file
	- launch the batch job with the Illumina DRAGEN batch json
	- a system loop looks for completion of the job to end the task.

Sample WDL: 
	
```
workflow myWorkflow {
    call myTask
}

task myTask {

    command {
        gsutil cp gs://jarvice-dragen-batch/batchsubmit.json .
        NAME=illumina-$(date +%Y%m%d%H%M)
        gcloud batch jobs submit $NAME --location us-central1 --config batchsubmit.json
        sleep 120
        echo "Batch $NAME started:"

        while true; do
                # Run your command here.
                output=`gcloud batch jobs describe --location us-central1 $NAME | grep state: | cut -c 10-20`
                if [[ "$output" == "SUCCEEDED" ]]; then
                echo "Job $NAME status: SUCCEEDED"
                break
                elif [[ "$output" == "FAILED" ]]; then
                echo "Job $NAME status: FAILED"
                break
                else
                # Continue looping.
                gcloud batch jobs describe --location us-central1 $NAME | grep state:
                sleep 60
                fi
        done

        echo "Batch job $NAME completed"
         }
    
    runtime {
        docker: "gcr.io/google.com/cloudsdktool/cloud-sdk:latest"
        memory: "2 GB"
        cpu: "1"
        time: "1:30:0"
        }

    output {
        String out = read_string(stdout())
    }
}
```

5. Launch the job by using the command:

`java -Dconfig.file=gcpbatchillumina.conf -jar cromwell-87-cc50a1f-SNAP.jar run cromwell-dragen.wdl`

We use the cromwell-87 compiled jar file. You can download the Cromwell page: https://github.com/broadinstitute/cromwell/releases
GCP Batch integration was not part of the standard cromwell release yet. You may need to compile the cromwell from the development branch.

Happy Computing
