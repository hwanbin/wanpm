package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hwanbin/wanpm-api/internal/s3action"
)

func (app *application) createPresignedPutUrlHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	fileName := app.readString(qs, "filename", "")
	if fileName == "" {
		app.badRequestResponse(w, r, fmt.Errorf("empty filename"))
	}

	lifetimeSecs := 60
	presigner := app.s3actor.presignClient

	request, err := presigner.PresignPutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket: aws.String(app.config.s3.bucket),
			Key:    aws.String(fileName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(lifetimeSecs) * time.Second
		},
	)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Couldn't get a presigned request to put %s: %v", fileName, err))
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"presigned": request}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createPresignedGetUrlHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	fileName := app.readString(qs, "filename", "")
	if fileName == "" {
		app.badRequestResponse(w, r, fmt.Errorf("empty filename"))
	}

	lifetimeSecs := 60
	presigner := app.s3actor.presignClient

	request, err := presigner.PresignGetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(app.config.s3.bucket),
			Key:    aws.String(fileName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(lifetimeSecs) * time.Second
		},
	)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Couldn't get a presigned request to get %s: %v", fileName, err))
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"presigned": request}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createPresignedDeleteUrlHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	fileName := app.readString(qs, "filename", "")
	if fileName == "" {
		app.badRequestResponse(w, r, fmt.Errorf("empty filename"))
	}

	lifetimeSecs := 60
	presigner := app.s3actor.presignClient

	request, err := presigner.PresignDeleteObject(
		context.Background(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(app.config.s3.bucket),
			Key:    aws.String(fileName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(lifetimeSecs) * time.Second
		},
	)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Couldn't get a presigned request to delete %s: %v", fileName, err))
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"presigned": request}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listFilesWithPrefixHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	prefix := app.readString(qs, "prefix", "")
	bucket := app.config.s3.bucket

	var fileNames []string

	fileNames, err := s3action.ListObjects(app.s3actor.client, bucket, prefix)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("unable to list objects with prefix %q: %v", prefix, err))
	}

	app.writeJSON(
		w,
		http.StatusOK,
		envelope{
			"base_url":   fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucket, "us-east-1"),
			"file_names": fileNames,
		},
		nil,
	)
}
