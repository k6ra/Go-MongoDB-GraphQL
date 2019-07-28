package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/graphql-go/graphql"
)

type User struct {
	Id       bson.ObjectId `bson:"_id"`
	Name     string        `bson:"name"`
	Email    string        `bson:"email"`
	Password string        `bson:"password"`
}

func main() {
	session, _ := mgo.Dial("mongodb://localhost/test")
	defer session.Close()
	db := session.DB("test")
	var AllUsers []User
	query := db.C("users").Find(bson.M{})
	query.All(&AllUsers)

	var userType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "User",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
				"email": &graphql.Field{
					Type: graphql.String,
				},
				"password": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)
	fields := graphql.Fields{
		"user": &graphql.Field{
			Type:        userType,
			Description: "Fetch user by Id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(param graphql.ResolveParams) (interface{}, error) {
				id, ok := param.Args["id"].(string)
				if ok {
					for _, user := range AllUsers {
						if user.Id.Hex() == id {
							return user, nil
						}
					}
				}
				return nil, nil
			},
		},
		"list": &graphql.Field{
			Type:        graphql.NewList(userType),
			Description: "Fetch users list",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return AllUsers, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: fields,
	}
	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	request := `
		{
			list {
				id
				name
				email
				password
			}
		}
	`
	params := graphql.Params{
		Schema:        schema,
		RequestString: request,
	}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON)
}
