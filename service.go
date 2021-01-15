package main

import (
	"ad-manager/pb"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	logger        log.Logger
	db            *gorm.DB
	storageClient *storage.Client
	grpcConn      *grpc.ClientConn
	requestId     int64
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

func MakeService(logger log.Logger, db *gorm.DB, storageClient *storage.Client, grpcConn *grpc.ClientConn) Service {
	db.AutoMigrate(&Ad{}, &Photo{})
	return &adService{
		logger:        log.With(logger, "component", "service"),
		db:            db,
		storageClient: storageClient,
		grpcConn:      grpcConn,
	}
}

func (s adService) PostAd(ctx context.Context, ad Ad) (uint, error) {
	logger := log.With(s.logger, "request-id", time.Now().UnixNano())

	logContext, _ := json.Marshal(ad)
	level.Info(logger).Log("msg", "PostAd request received", "context", logContext)
	ad.IdAd = 0
	if ad.IdUser == "" ||
		ad.Description == "" ||
		ad.Title == "" ||
		ad.Price == 0 {
		return 0, ErrMissingFields
	}
	result := s.db.Create(&ad)

	if result.Error != nil {
		level.Error(logger).Log("context", "PostAd", "msg", result.Error)
		return 0, result.Error
	}
	return ad.IdAd, nil
}

func (s adService) PutAd(ctx context.Context, ad Ad) error {
	logger := log.With(s.logger, "request-id", time.Now().UnixNano())

	logContext, _ := json.Marshal(ad)
	level.Info(logger).Log("msg", "PutAd request received", "context", logContext)

	ad.IdUser = ""
	if ad.IdAd == 0 {
		level.Error(logger).Log("context", "PutAd", "msg", ErrMissingFields)
		return ErrMissingFields
	}
	result := s.db.Model(&ad).Updates(ad)
	if result.RowsAffected < 1 {
		level.Error(logger).Log("context", "PutAd", "msg", ErrNotFound)
		return ErrNotFound
	}

	if result.Error != nil {
		level.Error(logger).Log("context", "PutAd", "msg", result.Error)
		return result.Error
	}
	return nil
}

func (s adService) DeleteAd(ctx context.Context, id uint) error {
	logger := log.With(s.logger, "request-id", time.Now().UnixNano())

	level.Info(logger).Log("msg", "DeleteAd request received", "context", fmt.Sprintf("\"id\":%d", id))

	result := s.db.Delete(&Ad{}, id)
	if result.RowsAffected < 1 {
		level.Error(logger).Log("context", "DeleteAd", "msg", ErrNotFound)
		return ErrNotFound
	}

	if result.Error != nil {
		level.Error(logger).Log("context", "DeleteAd", "msg", result.Error)
		return result.Error
	}
	return nil
}

func (s adService) PostPhoto(ctx context.Context, adId uint, file multipart.File) (*Photo, error) {
	requestId := fmt.Sprint(time.Now().UnixNano())
	logger := log.With(s.logger, "request-id", requestId)

	level.Info(logger).Log("msg", "PostPhoto request received", "context", fmt.Sprintf("\"id\":%d", adId))

	defer file.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	bucketName := "meshetr-images"
	objectName := fmt.Sprintf("%d-%d", adId, time.Now().UnixNano())
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	writer := s.storageClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		level.Error(logger).Log("context", "PostPhoto", "msg", ErrUpload)
		return nil, ErrUpload
	}
	if err := writer.Close(); err != nil {
		level.Error(logger).Log("context", "PostPhoto", "msg", err)
		return nil, err
	}

	photo := Photo{IdAd: adId, UrlOriginal: url}
	result := s.db.Create(&photo)
	if result.Error != nil {
		level.Error(logger).Log("context", "PostPhoto", "msg", result.Error)
		return nil, result.Error
	}

	client := pb.NewImageProcessorServiceClient(s.grpcConn)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, "request-id", requestId)
	defer cancel()
	status, err := client.Process(ctx, &pb.Image{Id: uint32(photo.IdPhoto)})
	if err != nil {
		level.Error(logger).Log("context", "PostPhoto", "msg", err)
		return nil, err
	}
	if status.Code != pb.StatusCode_Ok {
		level.Error(logger).Log("context", "PostPhoto", "msg", fmt.Sprintf("Received non 200 response code: %d", status.Code))
		return nil, errors.New(status.Message)
	}

	return &photo, nil
}

func (s adService) DeletePhoto(ctx context.Context, adId uint, id uint) error {
	logger := log.With(s.logger, "request-id", time.Now().UnixNano())

	level.Info(logger).Log("msg", "DeletePhoto request received", "context", fmt.Sprintf("\"id\":%d", id))
	result := s.db.Delete(&Photo{IdPhoto: id, IdAd: adId})
	if result.RowsAffected < 1 {
		level.Error(logger).Log("context", "DeletePhoto", "msg", ErrNotFound)
		return ErrNotFound
	}

	if result.Error != nil {
		level.Error(logger).Log("context", "DeletePhoto", "msg", result.Error)
		return result.Error
	}
	return nil
}
