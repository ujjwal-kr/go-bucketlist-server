package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber/v2"
	"github.com/ujjwal-kr/go-bucketlist-server/list"
	"github.com/ujjwal-kr/go-bucketlist-server/task"
	"github.com/ujjwal-kr/go-bucketlist-server/user"
)

// MongoInstance contains the Mongo client and database objects

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "mybucketlist"
const mongoURI = "mongodb://localhost:27017/" + dbName

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

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}

func setupRoutes(app *fiber.App) {

	app.Get("/", welcome)

	app.Get("/users", user.GetUsers)
	app.Get("/users/:id", user.GetUser) // gets the lists of the user along with the user
	app.Post("/auth/login", user.Login)
	app.Post("/auth/register", user.Register)
	app.Get("/users/:id/tasks", user.GetTasks)

	app.Post("/lists", list.PostList)
	app.Get("/lists/:id", list.GetList)
	app.Patch("/lists/:id", list.UpdateList)
	app.Delete("/lists/:id", list.DeleteList)

	app.Post("/tasks", task.PostTask)
	app.Delete("/tasks/:id", task.DeleteTask)

}

func main() {
	app := fiber.New()

	setupRoutes(app)

	app.Listen(":8080")
}
