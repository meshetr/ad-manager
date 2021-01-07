package main

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"time"
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrMissingFields   = errors.New("missing fields")
	ErrUpload          = errors.New("upload failed")
)

type Service interface {
	// Ad methiods
	PostAd(ctx context.Context, ad Ad) (uint, error)
	PutAd(ctx context.Context, ad Ad) error
	DeleteAd(ctx context.Context, id uint) error
	// Photo methods
	PostPhoto(ctx context.Context, adId uint, file multipart.File) (string, error)
	DeletePhoto(ctx context.Context, adId uint, id uint) error
}

type adService struct {
	db            *gorm.DB
	storageClient *storage.Client
}

type Ad struct {
	ID          uint    `json:"id"`
	UserId      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Photos      []Photo `json:"photos,omitempty"`
}

type Photo struct {
	ID          uint   `json:"id"`
	AdID        uint   `json:"ad_id"`
	Ad          Ad     `json:"-"`
	UrlOriginal string `json:"url_original"`
}

func MakeService(db *gorm.DB, storageClient *storage.Client) Service {
	db.AutoMigrate(&Ad{}, &Photo{})
	return &adService{
		db:            db,
		storageClient: storageClient,
	}
}

func (s adService) PostAd(ctx context.Context, ad Ad) (uint, error) {
	ad.ID = 0
	if ad.UserId == "" ||
		ad.Description == "" ||
		ad.Title == "" {
		return 0, ErrMissingFields
	}
	result := s.db.Create(&ad)
	return ad.ID, result.Error
}

func (s adService) PutAd(ctx context.Context, ad Ad) error {
	ad.UserId = ""
	if ad.ID == 0 ||
		ad.Description == "" ||
		ad.Title == "" {
		return ErrMissingFields
	}
	result := s.db.Model(&ad).Updates(ad)
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}

func (s adService) DeleteAd(ctx context.Context, id uint) error {
	result := s.db.Delete(&Ad{}, id)
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}

func (s adService) PostPhoto(ctx context.Context, adId uint, file multipart.File) (string, error) {
	defer file.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	bucketName := "meshetr-images"
	objectName := fmt.Sprintf("%d-%d", adId, time.Now().UnixNano())
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	writer := s.storageClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		return "", ErrUpload
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	result := s.db.Create(&Photo{AdID: adId, UrlOriginal: url})
	return url, result.Error
}

func (s adService) DeletePhoto(ctx context.Context, adId uint, id uint) error {
	result := s.db.Delete(&Photo{ID: id, AdID: adId})
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}
