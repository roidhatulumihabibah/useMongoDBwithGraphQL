package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Koneksi MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	mongoClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Membuat schema GraphQL
	fields := graphql.Fields{
		"data": &graphql.Field{
			Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
				Name: "MongoDBData",
				Fields: graphql.Fields{
					"name":  &graphql.Field{Type: graphql.String},
					"email": &graphql.Field{Type: graphql.String},
				},
			})),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var data []MongoDBData

				collection := mongoClient.Database("dosen").Collection("person")
				cur, err := collection.Find(context.Background(), nil)
				if err != nil {
					return nil, err
				}
				defer cur.Close(context.Background())

				for cur.Next(context.Background()) {
					var d MongoDBData
					err := cur.Decode(&d)
					if err != nil {
						return nil, err
					}
					data = append(data, d)
				}

				return data, nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Handler GraphQL
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Mengatur rute GraphQL
	http.Handle("/graphql", h)

	// Menjalankan server
	fmt.Println("Server GraphQL berjalan di http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
