package s3action

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func ListObjects(client *s3.Client, bucket, prefix string) ([]string, error) {
	var fileName []string

	paginator := s3.NewListObjectsV2Paginator(
		client,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		},
	)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			fileName = append(fileName, *obj.Key)
		}
	}

	return fileName, nil
}

func DeleteObjects(ctx context.Context, client *s3.Client, bucket string, objects []types.ObjectIdentifier) error {
	if len(objects) == 0 {
		return nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	}

	delOut, err := client.DeleteObjects(ctx, input)
	if err != nil || len(delOut.Errors) > 0 {
		log.Printf("Error deleting objects from bucket %s.\n", bucket)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", bucket)
				err = noBucket
			}
		} else if len(delOut.Errors) > 0 {
			for _, outErr := range delOut.Errors {
				log.Printf("%s: %s\n", *outErr.Key, *outErr.Message)
			}
			err = fmt.Errorf("%s", *delOut.Errors[0].Message)
		}
	} else {
		for _, delObjs := range delOut.Deleted {
			fmt.Printf("seeking '%s'", *delObjs.Key)
			err = s3.NewObjectNotExistsWaiter(client).Wait(
				ctx,
				&s3.HeadObjectInput{
					Bucket: aws.String(bucket),
					Key:    delObjs.Key,
				},
				time.Minute,
			)
			if err != nil {
				log.Printf("Failed attempt to wait for object %s to be deleted.\n", *delObjs.Key)
			} else {
				log.Printf("Deleted %s.\n", *delObjs.Key)
			}
		}
	}
	return err
}

func DeleteObject(ctx context.Context, client *s3.Client, bucket, key, versionId string) (bool, error) {
	deleted := false
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if versionId != "" {
		input.VersionId = aws.String(versionId)
	}

	_, err := client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &noKey) {
			log.Printf("Object %s does not exist in %s.\n", key, bucket)
			err = noKey
		} else if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				log.Printf("Access denied: cannot delete object %s from %s.\n", key, bucket)
				err = nil
			default:
				log.Println("Unexpected error")
			}
		}
	} else {
		err = s3.NewObjectNotExistsWaiter(client).Wait(
			ctx,
			&s3.HeadObjectInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: aws.String(versionId),
			},
			time.Minute,
		)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s in bucket %s to be deleted.\n", key, bucket)
		} else {
			deleted = true
		}
	}
	return deleted, err
}

func ListAllVersions(client *s3.Client, bucket, prefix string) ([]types.ObjectVersion, []types.DeleteMarkerEntry, error) {
	var err error
	var output *s3.ListObjectVersionsOutput
	var versions []types.ObjectVersion
	var deleteMarkers []types.DeleteMarkerEntry

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	versionPaginator := s3.NewListObjectVersionsPaginator(client, input)
	for versionPaginator.HasMorePages() {
		output, err = versionPaginator.NextPage(context.Background())
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", bucket)
				err = noBucket
			}
			return nil, nil, fmt.Errorf("unable to list object versions, %v", err)
			// break
		} else {
			versions = append(versions, output.Versions...)
			deleteMarkers = append(deleteMarkers, output.DeleteMarkers...)
		}

	}
	return versions, deleteMarkers, err
}

func RestoreDeletedObjects(ctx context.Context, client *s3.Client, bucket, prefix string) error {
	_, deleteMarkers, err := ListAllVersions(client, bucket, prefix)
	if err != nil {
		return err
	}
	for _, deleteMarker := range deleteMarkers {
		_, restoringErr := DeleteObject(ctx, client, bucket, *deleteMarker.Key, *deleteMarker.VersionId)
		if restoringErr != nil {
			return restoringErr
		}
	}
	return nil
}

func PermanentlyDeleteObjects(ctx context.Context, client *s3.Client, bucket, prefix string) error {
	versions, deleteMarkers, err := ListAllVersions(client, bucket, prefix)
	if err != nil {
		return err
	}
	for _, version := range versions {
		DeleteObject(ctx, client, bucket, *version.Key, *version.VersionId)
	}
	for _, deleteMarker := range deleteMarkers {
		DeleteObject(ctx, client, bucket, *deleteMarker.Key, *deleteMarker.VersionId)
	}
	return nil
}
