package worker

import (
	"context"
	"log"
	"time"

	"github.com/s1m0n21/go-crontab/common"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type logger struct {
	client     *mongo.Client
	collection *mongo.Collection
	logs       chan *common.Log
	ootCommit  chan *common.LogBatch
}

var Logger *logger

func (l *logger) saveLogs(batch *common.LogBatch) {
	_, err := l.collection.InsertMany(context.TODO(), batch.Logs)
	if err != nil {
		log.Println(err)
	}
}

func (l *logger) writeLoop() {
	var batch *common.LogBatch
	var timer *time.Timer

	for {
		select {
		case logStream := <-l.logs:
			if batch == nil {
				batch = &common.LogBatch{}
				timer = time.AfterFunc(time.Duration(Config.Log.CommitTimeout)*time.Millisecond, func(batch *common.LogBatch) func() {
					return func() {
						l.ootCommit <- batch
					}
				}(batch))
			}

			batch.Logs = append(batch.Logs, logStream)
			if len(batch.Logs) >= Config.Log.BatchSize {
				timer.Stop()
				l.saveLogs(batch)
				batch = nil
			}
		case ootBatch := <-l.ootCommit:
			if ootBatch != batch {
				continue
			}
			l.saveLogs(ootBatch)
			batch = nil
		}
	}
}

func InitLogger() error {
	opts := options.Client().ApplyURI(Config.Mongo.URI)

	client, err := mongo.Connect(context.TODO(), opts.SetConnectTimeout(time.Duration(Config.Mongo.ConnectTimeout)*time.Millisecond))
	if err != nil {
		return err
	}

	Logger = &logger{
		client:     client,
		collection: client.Database("cron").Collection("log"),
		logs:       make(chan *common.Log, 1000),
		ootCommit:  make(chan *common.LogBatch, 1000),
	}

	go Logger.writeLoop()

	return nil
}

func (l *logger) Append(log *common.Log) {
	select {
	case l.logs <- log:
	default:
	}
}
