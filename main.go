package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// MongoInstance contains the Mongo client and database objects

// const dbName = "mybucketlist"
// const mongoURI = "mongodb://localhost:27017/" + dbName

const mongoURI = "mongodb+srv://ujjwal:secretpassword@mybucket.wnews.mongodb.net/list?retryWrites=true&w=majority"

// MongoInstance contains the Mongo client and database objects
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

// TYPES

type User struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	TaskCode  string `json:"taskCode"`
	EntryCode string `json:"entryCode"`
}

type Task struct {
	ID     string `json:"id,omitempty" bson:"_id,omitempty"`
	Text   string `json:"text"`
	Date   string `json:"date"`
	UserId string `json:"userid"`
}

type List struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Text        string `json:"text"`
	Description string `json:"description"`
	UserId      string `json:"userid"`
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
	app.Use(cors.New())

	app.Get("/", welcome)

	users := app.Group("/users")
	auth := app.Group("/auth")
	lists := app.Group("/lists")
	tasks := app.Group("/tasks")

	// Users Handlers

	users.Get("/", protected, getAllUsers)
	users.Get("/:name", getUser)
	users.Get("/:name/tasks", protected, getUserTasks)

	auth.Post("/login", login)
	auth.Post("/register", register)
	auth.Get("/check", protected, checkAuth)

	// Lists Handlers

	lists.Get("/:id", getList)
	lists.Post("/", protected, postList)
	lists.Delete("/", protected, deleteList)

	// Tasks Handlers

	tasks.Post("/", protected, postTask)
	tasks.Delete("/", protected, deleteTask)

	// app.Listen(":8080")
	port := os.Getenv("PORT")
	app.Listen(":" + port)
}

//	Auth Middlewares

var Key = []byte("secret")

func protected(c *fiber.Ctx) error {
	tokenString := string(c.Request().Header.Peek("authorization"))
	claims := jwt.MapClaims{}
	if len(tokenString) > 1 {
		// gets the token based on the token string and verifies the signature with a private key
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(Key), nil
		})
		if err != nil {
			return c.Status(403).SendString("UNAUTHORIZED")
		}
		if token.Valid {
			c.Locals("userid", claims["username"])
			return c.Next()
		}
	}
	return c.Status(403).SendString("UNAUTHORIZED")
}

// 	Users Func

func register(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	user := new(User)
	// Parse Body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	if user.EntryCode != "secret" {
		return c.Status(500).SendString("not allowed")
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
	user.EntryCode = ""
	insertionResult, err := collection.InsertOne(c.Context(), user)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Status(201).JSON(insertionResult)
}

func login(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	user := &User{}
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).SendString("INVALID")
	}

	query := bson.D{{Key: "username", Value: user.Username}}
	userRecord := collection.FindOne(c.Context(), &query)
	existingUser := &User{}

	userRecord.Decode(&existingUser)
	if len(existingUser.ID) < 1 {
		return c.Status(404).SendString("cant find user")
	}
	if existingUser.Password != user.Password {
		return c.Status(403).SendString("UNAUTHORIZED")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 2000).Unix()

	tokenString, err := token.SignedString(Key)
	if err != nil {
		return c.Status(403).SendString("UNAUTHORIZED")
	}
	return c.Status(201).JSON(&fiber.Map{
		"id":    existingUser.ID,
		"name":  existingUser.Username,
		"token": tokenString,
	})
}

// Gets all the users and sanitizes the sensitive info

func getAllUsers(c *fiber.Ctx) error {
	collection := mg.Db.Collection("users")
	query := bson.D{{}}
	cursor, err := collection.Find(c.Context(), &query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var records []User = make([]User, 0)
	// iterate the cursor and decode the values
	if err := cursor.All(c.Context(), &records); err != nil {
		return c.Status(404).SendString("There isnt any")
	}
	var users []User = make([]User, 0)
	for i, s := range records {
		s.Password = ""
		s.TaskCode = ""
		users = append(users, s)
		fmt.Println(i)
	}

	return c.JSON(users)
}

// Get the specific user with username and returns the lists

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
	listQuery := bson.D{{Key: "userid", Value: user.Username}}
	cursor, err := Listscollection.Find(c.Context(), &listQuery)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var lists []List = make([]List, 0)
	if err := cursor.All(c.Context(), &lists); err != nil {
		return c.Status(500).SendString("internal err")
	}
	user.Password = ""
	user.TaskCode = ""
	return c.Status(200).JSON(&fiber.Map{
		"user":  user,
		"lists": lists,
	})
}

// Gets the user and its task, but requires to input "taskCode" header of user

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

	taskQuery := bson.D{{Key: "userid", Value: user.Username}}
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

func checkAuth(c *fiber.Ctx) error {
	return c.JSON(&fiber.Map{"message": "Seems Alrighty"})
}

//	List Funcs

func postList(c *fiber.Ctx) error {
	collection := mg.Db.Collection("lists")
	list := new(List)

	// Parse Body
	if err := c.BodyParser(&list); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	fmt.Println(c.Locals("userid"), list.UserId)
	// Check Locals
	if c.Locals("userid") == list.UserId {
		list.ID = ""
		insertionResult, err := collection.InsertOne(c.Context(), list)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.Status(201).JSON(insertionResult)
	}
	return c.Status(403).SendString("UNAUTHORIZEDdd")
}

func getList(c *fiber.Ctx) error {
	collection := mg.Db.Collection("lists")
	listID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	query := bson.D{{Key: "_id", Value: listID}}
	listRecord := collection.FindOne(c.Context(), &query)
	list := &List{}
	listRecord.Decode(&list)
	if len(list.ID) < 1 {
		return c.Status(404).SendString("Not found")
	}
	return c.JSON(list)
}

func deleteList(c *fiber.Ctx) error {
	collection := mg.Db.Collection("lists")
	listID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	query := bson.D{{Key: "_id", Value: listID}}

	itemRecord := collection.FindOne(c.Context(), bson.D{{}})
	item := &List{}
	itemRecord.Decode(&item)

	if c.Locals("userid") != item.UserId {
		return c.Status(403).SendString("UNAUTHORIZED")
	}

	res, err := collection.DeleteOne(c.Context(), &query)
	if err != nil || res.DeletedCount < 1 {
		return c.Status(404).SendString("Not found")
	}
	return c.SendString("Deleted")
}

// Task Funcs

func postTask(c *fiber.Ctx) error {
	collection := mg.Db.Collection("tasks")
	task := &Task{}

	// Parse Body
	if err := c.BodyParser(&task); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	// Check Locals
	if c.Locals("userid") != task.UserId {
		return c.Status(403).SendString("UNAUTHORIZED")
	}
	task.ID = ""
	insertionResult, err := collection.InsertOne(c.Context(), task)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Status(201).JSON(insertionResult)
}

func deleteTask(c *fiber.Ctx) error {
	collection := mg.Db.Collection("tasks")
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(404).SendString("Not found")
	}

	itemRecord := collection.FindOne(c.Context(), bson.D{{}})
	item := &Task{}
	itemRecord.Decode(&item)

	if c.Locals("userid") != item.UserId {
		return c.Status(403).SendString("UNAUTHORIZED")
	}

	query := bson.D{{Key: "_id", Value: taskID}}
	res, err := collection.DeleteOne(c.Context(), &query)
	if err != nil || res.DeletedCount < 1 {
		return c.Status(404).SendString("Not found")
	}
	return c.SendString("Deleted")
}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}
