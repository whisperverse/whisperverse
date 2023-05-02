package queries

import (
	"context"
	"fmt"

	"github.com/EmissarySocial/emissary/model"
	upgrades "github.com/EmissarySocial/emissary/queries/upgrades"
	"github.com/benpate/derp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpgradeMongoDB(connectionString string, databaseName string, domain *model.Domain) error {

	const currentDatabaseVersion = 3
	const location = "queries.UpgradeMongoDB"

	// If we're already at the target database version, then skip any other work
	if domain.DatabaseVersion == currentDatabaseVersion {
		return nil
	}

	ctx := context.Background()
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))

	if err != nil {
		return derp.Wrap(err, "data.mongodb.New", "Error creating mongodb client")
	}

	if err := client.Connect(ctx); err != nil {
		return derp.Wrap(err, "data.mongodb.New", "Error connecting to mongodb Server")
	}

	session := client.Database(databaseName)

	fmt.Println("UPGRADING DATABASE...")

	if domain.DatabaseVersion < 1 {
		if err := upgrades.Version1(ctx, session); err != nil {
			return derp.Wrap(err, location, "Error upgrading database to version 1")
		}
	}

	if domain.DatabaseVersion < 2 {
		if err := upgrades.Version2(ctx, session); err != nil {
			return derp.Wrap(err, location, "Error upgrading database to version 2")
		}
	}

	if domain.DatabaseVersion < 3 {
		if err := upgrades.Version3(ctx, session); err != nil {
			return derp.Wrap(err, location, "Error upgrading database to version 3")
		}
	}

	// Mark the Domain as "upgraded"
	domainCollection := session.Collection("Domain")

	filter := bson.M{"_id": primitive.NilObjectID}
	update := bson.M{"$set": bson.M{"databaseVersion": currentDatabaseVersion}}

	if _, err := domainCollection.UpdateOne(ctx, filter, update); err != nil {
		return derp.Wrap(err, location, "Error updating domain record")
	}

	fmt.Println(".")
	fmt.Println("DONE.")
	return nil
}