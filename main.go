package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

const dbName = "Instadb"
const collectionUsers = "users"
const collectionPosts = "posts"

func GetMongoDbConnection() (*mongo.Client, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

func getMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {
	client, err := GetMongoDbConnection()

	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}

type User struct {
	_id      string `json:"userid,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type Post struct {
	_id       string `json:"postid,omitempty"`
	User_id   string `json:"userid,omitempty"`
	Caption   string `json:"caption,omitempty"`
	Image_url string `json:"url,omitempty"`
	PostedAt  time.Time
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func addUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	collection, err := getMongoDbCollection(dbName, collectionUsers)
	if err != nil {
		panic(err)

	}

	var user User
	json.NewDecoder(r.Body).Decode(&user)

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	user.Password = string(hash)
	res, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		panic(err)

	}

	response, err := json.Marshal(res)
	if err != nil {
		panic(err)

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	collection, err := getMongoDbCollection(dbName, collectionUsers)
	if err != nil {
		panic(err)
	}

	var filter bson.M = bson.M{}
	t := ps.ByName("uid")
	if t != "" {
		id := t
		objID, _ := primitive.ObjectIDFromHex(id)
		filter = bson.M{"_id": objID}
	}

	var results []bson.M
	projection := bson.D{{"password", 0}}

	cur, err := collection.Find(context.Background(), filter, options.Find().SetProjection(projection))
	defer cur.Close(context.Background())

	if err != nil {
		panic(err)

	}

	cur.All(context.Background(), &results)

	json, _ := json.Marshal(results)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func addPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	collection, err := getMongoDbCollection(dbName, collectionPosts)
	if err != nil {
		panic(err)

	}

	var post Post
	json.NewDecoder(r.Body).Decode(&post)
	post.PostedAt = time.Now().Local()
	res, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		panic(err)

	}

	response, err := json.Marshal(res)
	if err != nil {
		panic(err)

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func getPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	collection, err := getMongoDbCollection(dbName, collectionPosts)
	if err != nil {
		panic(err)
	}

	var filter bson.M = bson.M{}
	t := ps.ByName("pid")
	if t != "" {
		id := t
		objID, _ := primitive.ObjectIDFromHex(id)
		filter = bson.M{"_id": objID}
	}

	var results []bson.M

	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	if err != nil {
		panic(err)

	}

	cur.All(context.Background(), &results)

	json, _ := json.Marshal(results)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func getUserPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	collection, err := getMongoDbCollection(dbName, collectionPosts)
	if err != nil {
		panic(err)
	}

	var filter bson.M = bson.M{}
	t := ps.ByName("uid")
	if t != "" {
		id := t
		filter = bson.M{"user_id": id}
	}

	var results []bson.M
	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	if err != nil {
		panic(err)

	}

	cur.All(context.Background(), &results)

	json, _ := json.Marshal(results)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/users", addUser)
	router.GET("/users/:uid", getUser)
	router.POST("/posts", addPost)
	router.GET("/posts/:pid", getPost)
	router.GET("/post/users/:uid", getUserPost)

	log.Fatal(http.ListenAndServe("localhost:8081", router))
}
