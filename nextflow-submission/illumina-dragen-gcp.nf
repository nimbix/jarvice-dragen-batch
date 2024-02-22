#!/usr/bin/env nextflow
nextflow.enable.dsl=2

process illumina_dragen_batch {
    container = 'gcr.io/google.com/cloudsdktool/google-cloud-cli'
    script:
    '''
        gsutil cp gs://jarvice-dragen-batch/batchsubmit.json .
        NAME="illumina-$(date +%Y%m%d%H%M)"
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
    '''

    output:
    stdout
}

workflow {
        illumina_dragen_batch | view
}
