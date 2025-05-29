package utils

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
	// Create a new storage client
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	// Ensure the client is closed after use
	defer client.Close()
	// Check if the bucket exists
	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
    if err == nil {
        log.Println("Backend bucket exists")
        return nil
    }
	// If the bucket does not exist, create it
    if errors.Is(err, storage.ErrBucketNotExist) {
        log.Println("Creating backend bucket...")
		err := bucket.Create(ctx, projectID, &storage.BucketAttrs{});
        if  err != nil {
            return fmt.Errorf("failed to create bucket: %w", err)
        }
        log.Println("Backend bucket created")
        return nil
    }
    return fmt.Errorf("failed to get bucket attributes: %w", err)
}