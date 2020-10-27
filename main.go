package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	lists := app.Group("/lists")
	tasks := app.Group("/tasks")

	// Users Handlers

	users.Get("/", welcome)
	users.Get("/:id", welcome)
	users.Get("/:id/tasks", welcome)

	auth.Post("/login", welcome)
	auth.Post("/register", Register)

	// Lists Handlers

	lists.Get("/:id", welcome)
	lists.Post("/", welcome)
	lists.Delete("/", welcome)

	// Tasks Handlers

	tasks.Post("/", welcome)
	tasks.Delete("/", welcome)

	app.Listen(":8080")
}

// 	Users Func

func Register(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	user := new(User)
	// Parse Body
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	username := user.Username
	query := bson.D{{Key: "username", Value: username}}
	existingRecord := collection.FindOne(c.Context(), &query)

	existingUser := &User{}
	existingRecord.Decode(&existingUser)
	if username == existingUser.Username {
		return c.Status(500).SendString("not allowed")
	}
	user.ID = ""
	insertionResult, err := collection.InsertOne(c.Context(), user)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fmt.Println(insertionResult)
	return c.Status(201).SendString("Created")

}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}
