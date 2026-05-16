package tracemongo

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const ttlIndexName = "started_at_ttl"

var _ trace.Repo = (*mongoRepo)(nil)

func NewMongoRepo(ctx context.Context, db *mongo.Database, ttl time.Duration) (*mongoRepo, error) {
	collection := db.Collection("traces")

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "request_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	if err := ensureTTLIndex(ctx, db, collection, ttl); err != nil {
		return nil, err
	}

	return &mongoRepo{collection: collection}, nil
}

func ensureTTLIndex(ctx context.Context, db *mongo.Database, collection *mongo.Collection, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	expireSeconds := int32(ttl.Seconds())

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "started_at", Value: 1}},
		Options: options.Index().SetName(ttlIndexName).SetExpireAfterSeconds(expireSeconds),
	})
	if err == nil {
		return nil
	}

	var cmdErr mongo.CommandError
	if !stderrors.As(err, &cmdErr) || (cmdErr.Code != 85 && cmdErr.Code != 86) {
		return err
	}

	return db.RunCommand(ctx, bson.D{
		{Key: "collMod", Value: collection.Name()},
		{Key: "index", Value: bson.D{
			{Key: "name", Value: ttlIndexName},
			{Key: "expireAfterSeconds", Value: expireSeconds},
		}},
	}).Err()
}

type mongoRepo struct {
	collection *mongo.Collection
}

func (r *mongoRepo) Save(ctx context.Context, t trace.Trace) error {
	doc := toDocument(t)
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *mongoRepo) GetByRequestID(ctx context.Context, requestID trace.RequestID) (trace.Trace, error) {
	var doc traceDocument
	err := r.collection.FindOne(ctx, bson.M{"request_id": requestID.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return trace.Trace{}, trace.ErrNotFound
		}
		return trace.Trace{}, err
	}
	return fromDocument(doc), nil
}

func (r *mongoRepo) DeleteByRequestID(ctx context.Context, requestID trace.RequestID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"request_id": requestID.String()})
	return err
}
