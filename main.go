package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func main() {
	endpoint := "localhost:9000"
	accessKeyID := "user"
	secretAccessKey := "pwd"
	useSSL := false

	// Init
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%#v", minioClient)

	// Make bucket
	bucketName := "mi-test"
	tmpBucketName := "mi-test-tmp"
	location := "us-east-1"
	prefix := "tmp"

	if err := minioClient.MakeBucket(bucketName, location); err != nil {
		log.Println(err)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	}

	if err := minioClient.MakeBucket(tmpBucketName, location); err != nil {
		log.Println(err)
		exists, err := minioClient.BucketExists(tmpBucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", tmpBucketName)
		} else {
			log.Fatalln(err)
		}
	}

	bytesBuff := bytes.NewBuffer([]byte{})
	sourceInfoArray := []minio.SourceInfo{}
	num := 20000000
	for i := 0; i <= num; i++ {
		s := fmt.Sprintf("%v%v", RandStringBytesRmndr(20), "\n")
		bytesBuff.WriteString(s)

		if bytesBuff.Len() > 1024*1024*50 || i == num {
			reader := bufio.NewReader(bytesBuff)
			fileName := fmt.Sprintf("%v-%v", prefix, i)
			_, err := minioClient.PutObject(tmpBucketName, fileName, reader, -1, minio.PutObjectOptions{})
			if err != nil {
				log.Println(err)
			}
			bytesBuff.Reset()

			srcInfo := minio.NewSourceInfo(tmpBucketName, fileName, nil)
			fmt.Printf("%+v\n", srcInfo)
			sourceInfoArray = append(sourceInfoArray, srcInfo)
		}
	}
	dst, err := minio.NewDestinationInfo(bucketName, "ABC", nil, nil)
	if err != nil {
		log.Fatalf("err1: %v", err)
	}
	err = minioClient.ComposeObject(dst, sourceInfoArray)
	if err != nil {
		log.Fatalf("err2: %v", err)
	}

	objectsCh := make(chan string)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range minioClient.ListObjects(tmpBucketName, prefix, true, nil) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object.Key
		}
	}()

	for rErr := range minioClient.RemoveObjects(tmpBucketName, objectsCh) {
		fmt.Println("Error detected during deletion: ", rErr)
	}
	err = minioClient.RemoveBucket(tmpBucketName)
	if err != nil {
		fmt.Println(err)
	}
}
