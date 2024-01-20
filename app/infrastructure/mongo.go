package infrastructure

import (
	"context"
	"data-collector-api/config"
	"data-collector-api/domain/entities"
	"fmt"
	"github.com/getsentry/sentry-go"
	"go.elastic.co/apm/module/apmmongo/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

type DBMongo struct {
	Client *mongo.Client
}

var AppMongo *DBMongo

func init() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	var dbMongo DBMongo
	fmt.Println("URI MONGO:", config.Config.GetString("MONGO_URL"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.Config.GetString("MONGO_URL")), options.Client().SetMonitor(apmmongo.CommandMonitor()))

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	dbMongo.Client = client
	AppMongo = &dbMongo
}

func (m *DBMongo) InsertMongo(ctx context.Context, data interface{}) error {
	span := sentry.StartSpan(ctx, "InsertMongo")
	defer span.Finish()
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	ctx = span.Context()
	coll := m.Client.Database(config.Config.GetString("MONGO_DB")).Collection(config.Config.GetString("MONGO_COLLECTION"))
	log.Println(coll)
	result, err := coll.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return nil
}

func (m *DBMongo) GetDataByUserID(ctx context.Context, userID interface{}) (entities.UserActivity, error) {
	span := sentry.StartSpan(ctx, "GetDataByUserID")
	defer span.Finish()
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	ctx = span.Context()

	coll := m.Client.Database(config.Config.GetString("MONGO_DB")).Collection(config.Config.GetString("MONGO_COLLECTION"))
	filter := bson.D{{"userid", userID}}
	project := bson.D{{"userid", 1}}
	opts := options.FindOne().SetProjection(project)
	var result entities.UserActivity
	err := coll.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		return entities.UserActivity{}, err
	}

	return result, nil
}

func (m *DBMongo) UpdatePushDetail(ctx context.Context, data, userID interface{}, element string) error {
	span := sentry.StartSpan(ctx, "UpdatePushDetail")
	defer span.Finish()
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	ctx = span.Context()

	coll := m.Client.Database(config.Config.GetString("MONGO_DB")).Collection(config.Config.GetString("MONGO_COLLECTION"))

	fmt.Println("preparation push data : ", data)
	filter := bson.D{{"userid", userID}}
	change := bson.M{"$push": bson.M{element: data}}
	result, err := coll.UpdateOne(ctx, filter, change)
	if err != nil {
		return err
	}
	log.Println(result)
	return nil
}
