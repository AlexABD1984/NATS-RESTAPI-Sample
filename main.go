package main

import (
	//"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	gojsonschema "github.com/xeipuuv/gojsonschema"
	"fmt"
	"io/ioutil"
	"github.com/nats-io/go-nats"
	//"runtime"
	"os"
	"time"
)

//Schema validation data (draft 7)
//In real production and for better code style is better to store it in uri and cache it to reduce network latancy while it fetch from central location
var SchemaString = `{
"$id": "http://example.com/example.json",
"type": "object",
"additionalProperties": false,
"required": [
"ts",
"sender",
"message"
],
"definitions": {},
"$schema": "http://json-schema.org/draft-07/schema#",
"properties": {
"ts": {
"$id": "/properties/ts",
"type": "string",
"title": "The Ts Schema ",
"default": "",
"pattern": "^[0-9]{1,10}$",
"examples": [
"1530228282"
]
},
"sender": {
"$id": "/properties/sender",
"type": "string",
"title": "The Sender Schema ",
"default": "",
"examples": [
"testy-test-service"
]
},
"message": {
"$id": "/properties/message",
"type": "object",
"minProperties": 1,
"properties": {},
"additionalProperties": {
"type": "string",
"minItems": 1,
"description": "string values"
}
},
"sent-from-ip": {
"$id": "/properties/sent-from-ip",
"type": "string",
"format": "ipv4",
"default": "",
"examples": [
"1.2.3.4"
]
},
"priority": {
"$id": "/properties/priority",
"type": "integer",
"title": "The Priority Schema ",
"default": 0,
"examples": [
2
]
}
}
}`

type server struct {
	nc *nats.Conn
}
// Display API version
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get request recived");
	fmt.Fprint(w,"Unity Validation API Test v 0.2")
}

// create a new item
func (s server)  postHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post request recived");
	//get json from http request body in string type
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//Validate json input by defined schema
	schemaLoader := gojsonschema.NewStringLoader(SchemaString)
	documentLoader := gojsonschema.NewStringLoader(string(body))
	jsonValidationResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if jsonValidationResult.Valid() {
		fmt.Fprintln(w,"The json parameter is valid. ")
		// Create server connection
		natsConnection, _ := nats.Connect(nats.DefaultURL)
		fmt.Fprintln(w,"Connected to " + nats.DefaultURL)
		// Subscribe to subject
		fmt.Fprintln(w,"Publish message")
		// Simple Publisher
		natsConnection.Publish("message", []byte("Hello World"))
	} else {
		fmt.Fprint(w,"The json parameter is not valid. see errors :\n")
		//for _, desc := range result.Errors() {
		//	fmt.Printf("- %s\n", desc)
		//}
	}
}
//health check endpoint to ensure service is up
func (s server) healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HealthCheck request recived");
	fmt.Fprintln(w, "OK")
}

// main function to boot up everything
func main() {

	var s server
	var err error
	uri := os.Getenv("NATS_URI")
	fmt.Println("NAT_URI="+uri)

	//try to connect to NATS server (retry 5 times)
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(nats.DefaultURL)
		if err == nil {
			s.nc = nc
			break
		}
		fmt.Println("Waiting before connecting to NATS at:", uri)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatal("Error establishing connection to NATS:", err)
	}
	fmt.Println("Connected to NATS at:", s.nc.ConnectedUrl())

	//Define http routings
	router := mux.NewRouter()
	router.HandleFunc("/healthcheck", s.healthz).Methods("GET")
	router.HandleFunc("/api/v1/unitytestapi", indexHandler).Methods("GET")
	router.HandleFunc("/api/v1/unitytestapi", s.postHandler).Methods("POST")

	//Run Http listener
	fmt.Printf("Server version 0.2 is listening on port 80...")
	log.Fatal(http.ListenAndServe(":80", router))
}