package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go-clean-ddd-es-template/internal/domain/entities"
)

// MongoUserReadRepository implements UserReadRepository using MongoDB
type MongoUserReadRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoUserReadRepository creates a new MongoDB user read repository
func NewMongoUserReadRepository(client *mongo.Client, database, collection string) *MongoUserReadRepository {
	return &MongoUserReadRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// SaveUser saves a user to MongoDB
func (r *MongoUserReadRepository) SaveUser(ctx context.Context, user *entities.UserReadModel) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Set timestamp if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, user)
	return err
}

// GetUserByID retrieves a user by ID from MongoDB
func (r *MongoUserReadRepository) GetUserByID(ctx context.Context, userID string) (*entities.UserReadModel, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"user_id": userID, "deleted_at": bson.M{"$exists": false}}

	var user entities.UserReadModel
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email from MongoDB
func (r *MongoUserReadRepository) GetUserByEmail(ctx context.Context, email string) (*entities.UserReadModel, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"email": email, "deleted_at": bson.M{"$exists": false}}

	var user entities.UserReadModel
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ListUsers retrieves a list of users from MongoDB with pagination
func (r *MongoUserReadRepository) ListUsers(ctx context.Context, page, pageSize int) ([]*entities.UserReadModel, int64, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Filter out deleted users
	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)

	// Find options
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	// Execute query
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var users []*entities.UserReadModel
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates a user in MongoDB
func (r *MongoUserReadRepository) UpdateUser(ctx context.Context, user *entities.UserReadModel) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	user.UpdatedAt = time.Now()

	filter := bson.M{"user_id": user.UserID}
	update := bson.M{"$set": user}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteUser soft deletes a user in MongoDB
func (r *MongoUserReadRepository) DeleteUser(ctx context.Context, userID string) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	now := time.Now()
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// SaveEvent saves a user event to MongoDB
func (r *MongoUserReadRepository) SaveEvent(ctx context.Context, event *entities.UserEvent) error {
	eventsCollection := r.client.Database(r.database).Collection(r.collection + "_events")

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := eventsCollection.InsertOne(ctx, event)
	return err
}

// GetUserEvents retrieves events for a user from MongoDB
func (r *MongoUserReadRepository) GetUserEvents(ctx context.Context, userID string) ([]*entities.UserEvent, error) {
	eventsCollection := r.client.Database(r.database).Collection(r.collection + "_events")

	filter := bson.M{"user_id": userID}
	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := eventsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*entities.UserEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventsByType retrieves events by type from MongoDB
func (r *MongoUserReadRepository) GetEventsByType(ctx context.Context, eventType string) ([]*entities.UserEvent, error) {
	eventsCollection := r.client.Database(r.database).Collection(r.collection + "_events")

	filter := bson.M{"event_type": eventType}
	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := eventsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*entities.UserEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
