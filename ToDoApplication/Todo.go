package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const fileName = "todolist.json"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the TODO app")
}

type Task struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

func readTodoList() []Task {
	tasks := make([]Task, 0)
	if byteVal, fileReadErr := os.ReadFile(fileName); fileReadErr != nil {
		log.Fatalf("can't open the file %v", fileReadErr)
	} else {
		if len(byteVal) == 0 {
			log.Println("File is empty")
			return tasks
		}
		if err := json.Unmarshal(byteVal, &tasks); err != nil {
			log.Fatalf("can't decode the json file %v", err)
		}
	}
	return tasks
}

func writeTodoList(tasks []Task) {
	if taskJson, err := json.MarshalIndent(tasks, "", "\t"); err != nil {
		log.Fatalf("Couldn't encode the tasks to json %v\n", err)
	} else {
		if err = os.WriteFile(fileName, taskJson, 0644); err != nil {
			log.Fatalf("Can't write the json file %v\n", err)
		}
	}
}

func clearFileContent() {
	if err := os.Truncate(fileName, 0); err != nil {
		log.Fatalf("Failed to truncate: %v", err)
	}
}

func handlePutRequest(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}

	// Modify one of the fields of the json document
	tasks := readTodoList()
	tasks = append(tasks, newTask)
	// write the file
	writeTodoList(tasks)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		log.Fatalf("There was an error decoding the post request body %v\n", err)
	}
	// write the file
	writeTodoList(tasks)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	outPutString := ""
	switch r.Method {
	case "GET":
		// return the json document
		tasks := readTodoList()
		json.NewEncoder(w).Encode(tasks)
	case "PUT":
		handlePutRequest(w, r)
	case "POST":
		handlePostRequest(w, r)
	case "DELETE":
		clearFileContent()
	default:
		outPutString = "Invalid request method \n"
	}
	fmt.Fprintf(w, outPutString)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/todo", TodoHandler)

	r.Methods("GET", "PUT", "POST", "DELETE")

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}
