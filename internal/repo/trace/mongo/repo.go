package tracemongo

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ trace.Repo = (*mongoRepo)(nil)

func NewMongoRepo(ctx context.Context, db *mongo.Database) (*mongoRepo, error) {
	collection := db.Collection("traces")

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "request_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	return &mongoRepo{collection: collection}, nil
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
