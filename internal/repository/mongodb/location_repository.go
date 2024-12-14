package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
	"github.com/yusufatac/bitaksi-case-study/internal/repository"
)

type locationRepository struct {
	collection *mongo.Collection
}

// NewLocationRepository creates a new MongoDB location repository
func NewLocationRepository(db *mongo.Database) repository.LocationRepository {
	collection := db.Collection("driver_locations")

	// Create geospatial index
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(err) // In production, handle this error appropriately
	}

	return &locationRepository{
		collection: collection,
	}
}

func (r *locationRepository) SaveLocation(ctx context.Context, location *domain.DriverLocation) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"driver_id": location.DriverID}
	update := bson.M{
		"$set": bson.M{
			"location":  location.Location,
			"status":    location.Status,
			"timestamp": location.Timestamp,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *locationRepository) SaveLocations(ctx context.Context, locations []*domain.DriverLocation) error {
	operations := make([]mongo.WriteModel, len(locations))

	for i, loc := range locations {
		filter := bson.M{"driver_id": loc.DriverID}
		update := bson.M{
			"$set": bson.M{
				"location":  loc.Location,
				"status":    loc.Status,
				"timestamp": loc.Timestamp,
			},
		}
		operations[i] = mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := r.collection.BulkWrite(ctx, operations, opts)
	return err
}

func (r *locationRepository) FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error) {
	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lon, lat},
				},
				"$maxDistance": radius,
			},
		},
		"status": "active",
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(10)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locations []*domain.DriverLocation
	if err = cursor.All(ctx, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}
