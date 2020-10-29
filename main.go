package main

import (
	"context"
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
	ID     string `json:"id,omitempty" bson:"_id,omitempty"`
	Text   string `json:"text"`
	Date   string `json:"date"`
	UserId string `json:"userId"`
}

type List struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Text        string `json:"text"`
	Description string `json:"description"`
	UserId      string `json:"userId"`
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

	users.Get("/", getAllUsers)
	users.Get("/:name", getUser)
	users.Get("/:name/tasks", getUserTasks)

	auth.Post("/login", welcome)
	auth.Post("/register", register)

	// Lists Handlers

	lists.Get("/:id", welcome)
	lists.Post("/", postList)
	lists.Delete("/", welcome)

	// Tasks Handlers

	tasks.Post("/", welcome)
	tasks.Delete("/", welcome)

	app.Listen(":8080")
}

// 	Users Func

func register(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	user := new(User)
	// Parse Body
	if err := c.BodyParser(&user); err != nil {
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
	return c.Status(201).JSON(insertionResult)
}

func getAllUsers(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	query := bson.D{{}}
	cursor, err := collection.Find(c.Context(), &query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var users []User = make([]User, 0)

	// iterate the cursor and decode the values
	if err := cursor.All(c.Context(), &users); err != nil {
		return c.Status(404).SendString("There isnt any")
	}

	return c.JSON(users)
}

func getUser(c *fiber.Ctx) error {
	Userscollection := mg.Db.Collection("users")
	Listscollection := mg.Db.Collection("lists")
	username := c.Params("name")
	userQuery := bson.D{{Key: "username", Value: username}}

	userRecord := Userscollection.FindOne(c.Context(), &userQuery)
	user := &User{}
	userRecord.Decode(&user)
	if len(user.ID) < 1 {
		return c.Status(404).SendString("cant find user")
	}
	listQuery := bson.D{{Key: "userid", Value: user.ID}}
	cursor, err := Listscollection.Find(c.Context(), &listQuery)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var lists []List = make([]List, 0)
	if err := cursor.All(c.Context(), &lists); err != nil {
		return c.Status(500).SendString("internal err")
	}

	return c.Status(200).JSON(&fiber.Map{
		"user":  user,
		"lists": lists,
	})
}

func getUserTasks(c *fiber.Ctx) error {
	Userscollection := mg.Db.Collection("users")
	Taskscollection := mg.Db.Collection("tasks")
	username := c.Params("name")
	userQuery := bson.D{{Key: "username", Value: username}}

	userRecord := Userscollection.FindOne(c.Context(), &userQuery)
	user := &User{}
	userRecord.Decode(&user)
	if len(user.ID) < 1 {
		return c.Status(404).SendString("cant find user")
	}

	if string(c.Request().Header.Peek("taskCode")) != user.TaskCode {
		return c.Status(403).SendString("UNAUTHORIZED")
	}

	taskQuery := bson.D{{Key: "userid", Value: user.ID}}
	cursor, err := Taskscollection.Find(c.Context(), &taskQuery)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var tasks []Task = make([]Task, 0)
	if err := cursor.All(c.Context(), &tasks); err != nil {
		return c.Status(500).SendString("internal err")
	}

	return c.Status(200).JSON(&fiber.Map{
		"user":  user,
		"tasks": tasks,
	})
}

//	List Funcs

func postList(c *fiber.Ctx) error {
	collection := mg.Db.Collection("lists")
	list := new(List)

	// Parse Body
	if err := c.BodyParser(&list); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	list.ID = ""

	insertionResult, err := collection.InsertOne(c.Context(), list)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Status(201).JSON(insertionResult)
}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}
