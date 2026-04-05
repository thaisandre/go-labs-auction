package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions_test?authSource=admin"
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skipf("Skipping test: could not connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping test: MongoDB not available: %v", err)
	}

	database := client.Database("auctions_test")
	collection := database.Collection("auctions")
	defer collection.Drop(ctx)

	os.Setenv("AUCTION_DURATION", "3s")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := &AuctionRepository{
		Collection: collection,
	}

	auctionEntity := &auction_entity.Auction{
		Id:          "test-auction-auto-close",
		ProductName: "Test Product",
		Category:    "Electronics",
		Description: "A test product for auto-close verification",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	internalErr := repo.CreateAuction(ctx, auctionEntity)
	if internalErr != nil {
		t.Fatalf("Error creating auction: %v", internalErr.Message)
	}

	// Verify auction was created with Active status
	var result AuctionEntityMongo
	filter := bson.M{"_id": auctionEntity.Id}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		t.Fatalf("Error finding auction: %v", err)
	}

	if result.Status != auction_entity.Active {
		t.Fatalf("Expected auction status to be Active (%d), got %d", auction_entity.Active, result.Status)
	}

	t.Log("Auction created with Active status. Waiting for auto-close (3s + buffer)...")

	// Wait for the auction duration + a small buffer
	time.Sleep(4 * time.Second)

	// Verify auction status changed to Completed
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		t.Fatalf("Error finding auction after auto-close: %v", err)
	}

	if result.Status != auction_entity.Completed {
		t.Fatalf("Expected auction status to be Completed (%d) after duration, got %d", auction_entity.Completed, result.Status)
	}

	t.Log("Auction was automatically closed successfully!")
}

func TestGetAuctionDuration(t *testing.T) {
	// Test default duration (no env var set)
	os.Unsetenv("AUCTION_DURATION")
	duration := getAuctionDuration()
	if duration != 5*time.Minute {
		t.Errorf("Expected default duration of 5m, got %s", duration)
	}

	// Test custom duration
	os.Setenv("AUCTION_DURATION", "10s")
	defer os.Unsetenv("AUCTION_DURATION")
	duration = getAuctionDuration()
	if duration != 10*time.Second {
		t.Errorf("Expected duration of 10s, got %s", duration)
	}

	// Test invalid duration falls back to default
	os.Setenv("AUCTION_DURATION", "invalid")
	duration = getAuctionDuration()
	if duration != 5*time.Minute {
		t.Errorf("Expected default duration of 5m for invalid input, got %s", duration)
	}
}
