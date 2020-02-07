package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/xeipuuv/gojsonschema"
)

type DBConfig struct {
	user		string
	password	string
	dbname		string
}

type Database struct {
	DB *gorm.DB
}

type handlerContext struct {
	database	*Database
}

//db model
type Person struct {
	Id			string	`json:"id"`
	FirstName	string	`json:"firstName"`
	LastName	string	`json:"lastName"`
	Age			int		`json:"age"`
}

type createUpdatePerson struct {
	FirstName	string	`json:"firstName"`
	LastName	string	`json:"lastName"`
	Age			int		`json:"age"`
}

//TODO implement jsonschema validation
const personSchema = `
{
    "title": "PersonSchema",
    "type": "object",
    "$schema": "http://json-schema.org/schema#",
    "additionalProperties": false,
    "properties": {
        "firstName": {
			"type": "string",
			"minLength": 1
        },
        "lastName": {
			"type": "string",
			"minLength": 1
        },
		"age": {
			"type": "integer"
        }
    },
    "required": [
		"firstName",
		"lastName",
		"age"
    ]
}
`

func validatePeople(payload string) (errors []string) {
	schemaLoader := gojsonschema.NewStringLoader(personSchema)

	documentLoader := gojsonschema.NewStringLoader(string(payload))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        panic(err.Error())
	}

	if result.Valid() {
    	fmt.Printf("The document is valid\n")
    } else {
		// fmt.Printf("The document is not valid. see errors :\n")
		errors := []string{"The document is not valid. see errors :\n"}
        for _, err := range result.Errors() {
			errors = append(errors, fmt.Sprintf("- %s\n", err)) 
		}
		if len(errors) > 1 {
			return errors
		}
	}
	
	return nil
}

func main() {
	dbconfig := &DBConfig{user: "root", password: "mysql", dbname: "golang_api"}
	connectionUrl := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbconfig.user, dbconfig.password, dbconfig.dbname)

	db, err := gorm.Open("mysql",  connectionUrl)
	if err != nil {
		panic("failed to connect database")
	}

	context := &handlerContext{database: &Database{DB: db}}

	r := mux.NewRouter()
	r.HandleFunc("/people", context.PeoplePostHandler).Methods("POST")
	r.HandleFunc("/people", context.PeopleGetCollectionHandler).Methods("GET")
	r.HandleFunc("/people/{id:[0-9]+}", context.PeopleGetDetailHandler).Methods("GET")
	r.HandleFunc("/people/{id:[0-9]+}", context.PeoplePutHandler).Methods("PUT")
	r.HandleFunc("/people/{id:[0-9]+}", context.PeopleDeleteHandler).Methods("DELETE")
	http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}

func (context *handlerContext) PeoplePostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	schemaErrors := validatePeople(string(body))

	if schemaErrors != nil {
		http.Error(w, strings.Join(schemaErrors, ""), http.StatusBadRequest)
		return
	}

	person := createUpdatePerson{}

	err = json.Unmarshal([]byte(body), &person)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	newPerson := Person{
		FirstName: person.FirstName,
		LastName: person.LastName,
		Age: person.Age,
	}

	err = context.database.DB.Create(&newPerson).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	fmt.Fprintf(w, "succcess!")
}

func (context *handlerContext) PeoplePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	personID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	var personRecord Person

	err =  context.database.DB.Find(&personRecord, personID).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	schemaErrors := validatePeople(string(body))

	if schemaErrors != nil {
		http.Error(w, strings.Join(schemaErrors, ""), http.StatusBadRequest)
		return
	}

	person := createUpdatePerson{}

	err = json.Unmarshal([]byte(body), &person)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	update := false

	if person.FirstName != personRecord.FirstName {
		personRecord.FirstName = person.FirstName
		update = true
	}

	if person.LastName != personRecord.LastName {
		personRecord.LastName = person.LastName
		update = true
	}

	if person.FirstName != personRecord.FirstName {
		personRecord.LastName = person.LastName
		update = true
	}

	err = context.database.DB.Save(&personRecord).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	msg := ""
	if update == true {
		msg = "Updated!"
	} else {
		msg = "No change detected"
	}

	fmt.Fprintf(w, "%s", msg)
}

func (context *handlerContext) PeopleGetDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	personID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	var personRecord Person

	err = context.database.DB.Find(&personRecord, personID).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	payload, err := json.Marshal(personRecord)

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	fmt.Fprintf(w, "%s", payload)
}

func (context *handlerContext) PeopleGetCollectionHandler(w http.ResponseWriter, r *http.Request) {
	var personRecords []Person

	err := context.database.DB.Find(&personRecords).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	payload, err := json.Marshal(personRecords)

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	fmt.Fprintf(w, "%s", payload)
}

func (context *handlerContext) PeopleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	personID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	var personRecord Person

	err = context.database.DB.Delete(&personRecord, personID).Error

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	fmt.Fprintf(w, "succcess!")

}