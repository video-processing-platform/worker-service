package grpcclient

import (
	"context"
	"time"

	pb "github.com/alimarzban99/video-processor-service/proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	client pb.NotificationServiceClient
}

func NewNotificationClient() (*NotificationClient, error) {

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	return &NotificationClient{
		client: pb.NewNotificationServiceClient(conn),
	}, nil
}

func (n *NotificationClient) Send(
	userID string,
	title string,
	message string,
) error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_, err := n.client.SendEmail(
		ctx,
		&pb.SendEmailRequest{
			Email:   userID,
			Subject: title,
			Body:    message,
		},
	)

	return err
}

func (n *NotificationClient) SendProcessingCompleted(
	userID string,
	title string,
	message string,
) error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_, err := n.client.SendProcessingCompleted(
		ctx,
		&pb.ProcessingCompletedRequest{
			VideoId:  userID,
			Email:    userID,
			Filename: title,
			Status:   message,
		},
	)

	return err
}
