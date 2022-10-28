package handler

import (
	"Audiophile/database/helper"
	"Audiophile/utilities"
	"cloud.google.com/go/firestore"
	cloud "cloud.google.com/go/storage"
	"context"
	firebase "firebase.google.com/go"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type App struct {
	Ctx     context.Context
	Client  *firestore.Client
	Storage *cloud.Client
}

func Upload(r *http.Request) string {
	client := App{}
	client.Ctx = context.Background()
	credentialsFile := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
	// file :=
	var err error

	app, err := firebase.NewApp(client.Ctx, nil, credentialsFile)
	if err != nil {
		logrus.Printf("error is: %v", err)
		return ""
	}
	client.Client, err = app.Firestore(client.Ctx)
	if err != nil {
		logrus.Printf("error is:%v", err)
		return ""
	}
	client.Storage, err = cloud.NewClient(client.Ctx, credentialsFile)
	if err != nil {
		logrus.Printf("error is:%v", err)
		return ""
	}
	file, handler, err := r.FormFile("image")

	if err != nil {
		logrus.Printf("error is:%v", err)
		return ""
	}
	err1 := r.ParseMultipartForm(10 << 20)
	if err1 != nil {
		return ""
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			logrus.Printf("Upload: unable to close file")
			return
		}
	}(file)

	imagePath := handler.Filename

	bucket := "audiophile-84832.appspot.com"

	wc := client.Storage.Bucket(bucket).Object("images/" + imagePath).NewWriter(client.Ctx)
	_, err = io.Copy(wc, file)
	if err != nil {
		logrus.Printf("error is:%v", err)
		return ""
	}
	if err := wc.Close(); err != nil {
		logrus.Printf("error is:%v", err)
		return ""
	}

	URL := "https://storage.cloud.google.com/" + bucket + "/" + "images/" + imagePath

	return URL
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	url := Upload(r)

	imageID, err := helper.UploadImage(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UploadImage: not able to upload image:%v", err)
		return
	}
	err = utilities.Encoder(w, imageID)
	if err != nil {
		logrus.Printf("UploadImage: Encoder error:%v", err)
		return
	}
}
