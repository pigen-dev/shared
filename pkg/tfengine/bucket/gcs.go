package bucket

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
)

func SetupBackend(bucketName, projectID string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
			// TODO: Handle error.
	}
	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
	exist := true
	if err != nil {
		// Other error (e.g. permission issue)
		if !errors.Is(err, storage.ErrBucketNotExist) {
			return fmt.Errorf("failed to get bucket attributes: %w", err)
		}
		// If the bucket does not exist
		exist = false
	}
	if !exist{
		log.Println("Creating backend bucket...")
		err = bucket.Create(ctx, projectID, &storage.BucketAttrs{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	}
	log.Println("Vackend bucket exists")
	return nil
}