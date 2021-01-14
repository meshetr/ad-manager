package main

import (
	"ad-manager/pb"
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
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
	PostPhoto(ctx context.Context, adId uint, file multipart.File) (*Photo, error)
	DeletePhoto(ctx context.Context, adId uint, id uint) error
}

type adService struct {
	db            *gorm.DB
	storageClient *storage.Client
	grpcConn      *grpc.ClientConn
}

type Ad struct {
	IdAd        uint      `json:"id_ad" gorm:"primaryKey"`
	IdUser      string    `json:"id_user"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float32   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Ad) TableName() string {
	return "t_ad"
}

type Photo struct {
	IdPhoto     uint   `json:"id" gorm:"primaryKey"`
	IdAd        uint   `json:"id_ad"`
	Ad          Ad     `json:"-" gorm:"foreignKey:IdAd"`
	UrlOriginal string `json:"url_original"`
}

func (Photo) TableName() string {
	return "t_photo"
}

func MakeService(db *gorm.DB, storageClient *storage.Client, grpcConn *grpc.ClientConn) Service {
	db.AutoMigrate(&Ad{}, &Photo{})
	return &adService{
		db:            db,
		storageClient: storageClient,
		grpcConn:      grpcConn,
	}
}

func (s adService) PostAd(ctx context.Context, ad Ad) (uint, error) {
	ad.IdAd = 0
	if ad.IdUser == "" ||
		ad.Description == "" ||
		ad.Title == "" ||
		ad.Price == 0 {
		return 0, ErrMissingFields
	}
	result := s.db.Create(&ad)
	return ad.IdAd, result.Error
}

func (s adService) PutAd(ctx context.Context, ad Ad) error {
	ad.IdUser = ""
	if ad.IdAd == 0 {
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

func (s adService) PostPhoto(ctx context.Context, adId uint, file multipart.File) (*Photo, error) {
	defer file.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	bucketName := "meshetr-images"
	objectName := fmt.Sprintf("%d-%d", adId, time.Now().UnixNano())
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	writer := s.storageClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		return nil, ErrUpload
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	photo := Photo{IdAd: adId, UrlOriginal: url}
	result := s.db.Create(&photo)
	if result.Error != nil {
		return nil, result.Error
	}

	client := pb.NewImageProcessorServiceClient(s.grpcConn)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	status, err := client.Process(ctx, &pb.Image{Id: uint32(photo.IdPhoto)})
	if err != nil {
		return nil, err
	}
	if status.Code != pb.StatusCode_Ok {
		return nil, errors.New(status.Message)
	}

	return &photo, nil
}

func (s adService) DeletePhoto(ctx context.Context, adId uint, id uint) error {
	result := s.db.Delete(&Photo{IdPhoto: id, IdAd: adId})
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}
