package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
	"github.com/yusufatac/bitaksi-case-study/internal/repository"
)

type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(db *mongo.Database) repository.UserRepository {
	collection := db.Collection("users")

	// Create unique index for username
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "username", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(err) // In production, handle this error appropriately
	}

	return &userRepository{
		collection: collection,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateUsername
	}
	return err
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	filter := bson.M{
		"username": username,
		"status":   bson.M{"$ne": "deleted"},
	}

	var user domain.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"last_login_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Custom errors
var (
	ErrDuplicateUsername = errors.New("username already exists")
)
