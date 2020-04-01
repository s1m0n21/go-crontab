package master

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	logging "log"
	"time"

	"github.com/s1m0n21/go-crontab/common"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type logmgr struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var Logmgr *logmgr

func (lm *logmgr) ListLogs(name string, skip int, limit int) ([]*common.Log, error) {
	filter := common.LogFilter{JobName: name}
	//bFilter, err := bson.Marshal(&filter)
	//logging.Println(err)

	sort := &common.SortLogByStartTime{SortOrder: -1}

	bSort, err := bson.Marshal(sort)
	logging.Println(err)

	opts := options.Find()
	opts.SetSort(bSort)
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))

	cursor, err := lm.collection.Find(context.TODO(), filter, opts)
	defer cursor.Close(context.TODO())
	if err != nil {
		return nil, err
	}

	logArr := make([]*common.Log, 0)

	for cursor.Next(context.TODO()) {
		log := common.Log{}

		if err := cursor.Decode(&log); err != nil {
			logging.Println(err)
			continue
		}

		logArr = append(logArr, &log)
	}

	logging.Println(logArr)

	return logArr, nil
}

func InitLogmgr() error {
	opts := options.Client().ApplyURI(Config.Mongo.URI)

	client, err := mongo.Connect(context.TODO(), opts.SetConnectTimeout(time.Duration(Config.Mongo.ConnectTimeout)*time.Millisecond))
	if err != nil {
		return err
	}

	Logmgr = &logmgr{
		client:     client,
		collection: client.Database("cron").Collection("log"),
	}

	return nil
}
