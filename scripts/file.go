package scripts

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tajimyradov/transcoder/models"

	"github.com/minio/minio-go/v7"
)

func DownloadLargeObject(client *minio.Client, bucket, objectName, localFilePath string) error {
	ctx := context.Background()

	object, err := client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	stat, err := object.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat object: %w", err)
	}
	fmt.Printf("Downloading object (%s) - Size: %d bytes\n", objectName, stat.Size)

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	buf := make([]byte, 10*1024*1024)

	var totalWritten int64
	for {
		n, readErr := object.Read(buf)
		if n > 0 {
			written, writeErr := localFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}
			totalWritten += int64(written)
			fmt.Printf("\rDownloaded %d bytes...", totalWritten)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read error: %w", readErr)
		}
	}

	fmt.Println("\nDownload finished successfully.")
	return nil
}

func ClearFiles(input string) error {
	return os.RemoveAll(input)
}

func UploadFiles(filePath string, minioClient *minio.Client, config models.AppConfig) error {

	exists, err := minioClient.BucketExists(context.Background(), config.Minio.HLSFileBucket)
	if err != nil {
		return err
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), config.Minio.HLSFileBucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	// Folder upload
	err = filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		objectName, err := filepath.Rel(filePath, path)
		if err != nil {
			return err
		}

		if objectName == "." {
			objectName = info.Name()
		}

		objectName = strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1] + "/" + objectName

		_, err = minioClient.PutObject(
			context.Background(),
			config.Minio.HLSFileBucket,
			objectName,
			file,
			info.Size(),
			minio.PutObjectOptions{ContentType: "application/octet-stream"},
		)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
