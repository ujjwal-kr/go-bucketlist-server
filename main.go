package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber/v2"
)

// MongoInstance contains the Mongo client and database objects

const dbName = "mybucketlist"
const mongoURI = "mongodb://localhost:27017/" + dbName

// MongoInstance contains the Mongo client and database objects
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

// TYPES

type User struct {
	ID       string `json:"id,omitempty" bson:"_id,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
	TaskCode string `json:"taskCode"`
}

type Task struct {
	ID   string `json:"id,omitempty" bson:"_id,omitempty"`
	Text string `json:"text"`
	Date string `json:"date"`
}

type List struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Text        string `json:"text"`
	Description string `json:"description"`
}

var mg MongoInstance

// Connect configures the MongoDB client and initializes the database connection.
func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	return nil
}

func main() {

	// Connect to the database
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Get("/", welcome)

	users := app.Group("/users")
	auth := app.Group("/auth")

	// Users Handlers
	users.Get("/", welcome)
	users.Get("/:id", welcome)
	users.Get("/:id/tasks", welcome)

	auth.Post("/login", welcome)
	auth.Post("register", welcome)

	app.Listen(":8080")
}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}
