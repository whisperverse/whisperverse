package config

import (
	"context"
	"os"

	"github.com/benpate/derp"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage is a MongoDB-backed configuration storage
type MongoStorage struct {
	source        string
	location      string
	collection    *mongo.Collection
	updateChannel chan Config
}

// NewMongoStorage creates a fully initialized MongoStorage instance
func NewMongoStorage(args *CommandLineArgs) MongoStorage {

	// Create a new MongoDB database connection
	connectOptions := options.Client().ApplyURI(args.Location)
	client, err := mongo.Connect(context.Background(), connectOptions)

	if err != nil {
		log.Fatal().Msg("Emissary cannot start because the MongoDB config database could not be reached.")
		log.Fatal().Msg("Check the MongoDB connection string and verify the database server connection.")
		log.Error().Err(err).Send()
		os.Exit(1)
	}

	// Get the configuration collection
	collection := client.Database("emissary").Collection("config")

	storage := MongoStorage{
		source:        args.Source,
		location:      args.Location,
		collection:    collection,
		updateChannel: make(chan Config, 1),
	}

	// Special rules for the first time we load the configuration file
	config, err := storage.load()

	switch {

	// If the config was read successfully, then NOOP here skips down to the next section.
	case err == nil:

	case derp.NotFound(err):

		// If the config was not found, then run in setup mode and add a new default configuration
		args.Setup = true

		// Create a default configuration
		config = DefaultConfig()
		config.Source = storage.source
		config.Location = storage.location

		if err := storage.Write(config); err != nil {
			log.Fatal().Msg("Error writing new configuration file to the Mongo database")
			log.Fatal().Err(err).Send()
			os.Exit(1)
		}

	default:
		// Any other errors connecting to the Mongo server will prevent Emissary from starting.
		log.Fatal().Msg("Emissary could not start because of an error connecting to the MongoDB config database.")
		log.Fatal().Err(err).Send()
		os.Exit(1)
	}

	// If we have a valid config, post it to the update channel
	storage.updateChannel <- config

	// After the first load, watch for changes to the config record and post them to the update channel
	go func() {

		// watch for changes to the configuration
		cs, err := storage.collection.Watch(context.Background(), mongo.Pipeline{})

		if err != nil {
			derp.Report(derp.Wrap(err, "service.Watcher", "Unable to open Mongodb Change Stream"))
			return
		}

		for cs.Next(context.Background()) {
			if config, err := storage.load(); err == nil {
				storage.updateChannel <- config
			} else {
				derp.Report(derp.Wrap(err, "config.MongoStorage", "Error loading updated config from MongoDB"))
			}
		}
	}()

	return storage
}

// Subscribe returns a channel that will receive the configuration every time it is updated
func (storage MongoStorage) Subscribe() <-chan Config {
	return storage.updateChannel
}

// load reads the configuration from the MongoDB database
func (storage MongoStorage) load() (Config, error) {

	result := NewConfig()

	if err := storage.collection.FindOne(context.Background(), bson.M{}).Decode(&result); err != nil {
		return Config{}, derp.Wrap(err, "config.MongoStorage", "Error decoding config from MongoDB")
	}

	result.Source = storage.source
	result.Location = storage.location

	return result, nil
}

// Write writes the configuration to the database
func (storage MongoStorage) Write(config Config) error {

	upsert := true
	criteria := bson.M{"_id": config.MongoID}
	options := options.ReplaceOptions{
		Upsert: &upsert,
	}

	if _, err := storage.collection.ReplaceOne(context.Background(), criteria, config, &options); err != nil {
		return derp.Wrap(err, "config.MongoStorage", "Error writing config to MongoDB")
	}

	return nil
}
