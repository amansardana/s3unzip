package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const tempFolder string = "/tmp/unzip-test/"
const tempDownloadFolder string = "/tmp/unzip-test-download/"

func s3Unzip(downloadBucket, uploadBucket, item string) error{
	filePath := filepath.Join(tempDownloadFolder, item)
	optFolder := filepath.Join(tempFolder, strings.Replace(item, ".zip", "", 1))

	// Download
	if err:=s3Download(downloadBucket, item, tempDownloadFolder);err!=nil {
		return err
	}

	fmt.Println("Downloaded Successfully")
	// Unzip
	files, err := Unzip(filePath, optFolder)
	if err != nil {
		return err
	}

	fmt.Println("Unzipped:\n" + strings.Join(files, "\n"))

	// Upload
	if err:=s3Upload(uploadBucket, optFolder);err!=nil {
		return err
	}
	return nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func s3Download(bucket, item, dirPath string) error{
	// NOTE: you need to store your AWS credentials in ~/.aws/credentials

	// 2) Create an AWS session
	sess := makeSession()

	// 3) Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	os.MkdirAll(dirPath, os.ModePerm)
	f := filepath.Join(dirPath, item)
	file, err := os.Create(f)
	if err != nil {
		log.Fatalf("Unable to open file %v", err)
	}

	defer file.Close()

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		return fmt.Errorf("Unable to download item %q, %v", item, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
	return nil
}

func s3Upload(bucket, path string) error{

	// Make an aws session
	sess := makeSession()
	var err error
	if isDirectory(path) {
		err=uploadDirToS3(sess, bucket, path)
	} else {
		err=uploadFileToS3(sess, bucket, path)
	}
	return err
}

func isDirectory(path string) bool {
	fd, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	switch mode := fd.Mode(); {
	case mode.IsDir():
		return true
	case mode.IsRegular():
		return false
	}
	return false
}

func uploadDirToS3(sess *session.Session, bucketName string, dirPath string) error{
	fileList := []string{}
	filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		fmt.Println("PATH ==> " + path)
		if isDirectory(path) {
			// Do nothing
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})

	for _, file := range fileList {
		if err:=uploadFileToS3(sess, bucketName, file);err!=nil {
			return err
		}
	}
	return nil
}

func uploadFileToS3(sess *session.Session, bucketName string, filePath string) error{
	fmt.Println("upload " + filePath + " to S3")
	// An s3 service
	s3Svc := s3.New(sess)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file", file, err)
		os.Exit(1)
	}
	defer file.Close()
	var key string

	fileDirectory, _ := filepath.Abs(filePath)
	key = strings.Replace(fileDirectory, tempFolder, "", 1)

	// Upload the file to the s3 given bucket
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucketName), // Required
		Key:    aws.String(key),        // Required
		Body:   file,
	}
	_, err = s3Svc.PutObject(params)
	if err != nil {
		return fmt.Errorf("Failed to upload data to %s/%s, %s\n",
			bucketName, key, err.Error())
	}
	return nil
}

func makeSession() *session.Session {
	// Enable loading shared config file
	// Specify profile to load for the session's config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Println("failed to create session,", err)
		fmt.Println(err)
		os.Exit(1)
	}

	return sess
}
