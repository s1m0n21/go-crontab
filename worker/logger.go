package worker

import (
	"context"

	"github.com/s1m0n21/go-crontab/common"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type logger struct {
	client     *mongo.Client
	collection *mongo.Collection
	logChan    chan *common.Log
}

var Logger *logger

func InitLogger() error {
	cOpts := options.Client().ApplyURI("")

	client, err := mongo.Connect(context.TODO())
}
