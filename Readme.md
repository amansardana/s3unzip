# S3Unzip

It is a test program to unzip a zip file uploaded to a s3 bucket and upload unzip content to another bucket.

### Setup
To test you will need to setup golang and serverless environment.

Go Installation: https://golang.org/doc/install

AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html

Serverless Installation: `npm install -g serverless`

After you have go setup, use the following command:

**Install Dep**:
`curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh`

**Deploy**:
`make deploy`

### Explanation
It will create:
2 buckets named: `unzip-files.amansardana.com` and `zip-files.amansardana.com`
2 lambda functions: `unzip-test-dev-unzip` and `unzip-test-dev-s3unzip`
1 Api Gateway Endpoint: `https://<resourceID>.execute-api.us-east-1.amazonaws.com/dev/unzip`

On uploading a zip file to bucket `zip-files.amansardana.com` an `ObjectCreated:*` event will trigger `unzip-test-dev-s3unzip`
which will further trigger `unzip-files.amansardana.com` for each individual record received.

`unzip-files.amansardana.com` will download and unzip the file and upload unzip content to `unzip-files.amansardana.com`

**Endpoint:**

POST Request:
```
{
    "downloadBucket":"zip-files.amansardana.com",
    "uploadBucket":"unzip-files.amansardana.com",
    "item":"SampleZIPFile_50mbmb.zip"
}
```

**Sample Zip:**

https://www.sample-videos.com/download-sample-zip.php
