package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID   int
	Task string
	Done bool
}

type NewTodo struct {
	Task string
}

type app struct {
	router *mux.Router
	db     *sql.DB
	server *http.Server
}

func (a *app) initializeDb(dbPath string) error {
	var err error
	var statement *sql.Stmt

	if a.db, err = sql.Open("sqlite3", dbPath); err != nil {
		return err
	}

	if statement, err = a.db.Prepare("create table if not exists todo (id integer primary key, task string not null, done int default (0));"); err != nil {
		return err
	}

	if _, err = statement.Exec(); err != nil {
		return err
	}

	return err
}

func (a *app) initializeRoutes() {
	a.router = mux.NewRouter()
	a.router.HandleFunc("/todos", a.handleGetAllTodos).Methods("GET")  // list all todos
	a.router.HandleFunc("/todo", a.handleAddTodo).Methods("POST")      // Add a new todo
	a.router.HandleFunc("/todo/{id}", a.handleGetTodo).Methods("GET")  // fetch todo with a given id
	a.router.HandleFunc("/todo/{id}", a.handleTodoDone).Methods("PUT") // set todo with a given id
}

func (a *app) getAll() ([]Todo, error) {
	var tasks = make([]Todo, 0)

	rows, err := a.db.Query("select id, task, done from todo;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var task string
	var done int

	for rows.Next() {
		if err := rows.Scan(&id, &task, &done); err != nil {
			log.Printf("Could not read data from Todo table - %s", err)
		} else {
			if done == 0 {
				tasks = append(tasks, Todo{id, task, false})
			} else {
				tasks = append(tasks, Todo{id, task, true})
			}

		}
	}

	return tasks, nil
}

func (a *app) getTodo(taskID int) ([]Todo, error) {
	var tasks = make([]Todo, 0)

	rows, err := a.db.Query("select id, task, get from todo;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var task string
	var done int

	for rows.Next() {
		if err := rows.Scan(&id, &task, &done); err != nil {
			log.Printf("Could not read data from Todo table - %s", err)
		} else {
			if done == 0 {
				tasks = append(tasks, Todo{id, task, false})
			} else {
				tasks = append(tasks, Todo{id, task, true})
			}

		}
	}

	return tasks, nil
}

func (a *app) addTodo(task string) (int, error) {
	var err error
	var statement *sql.Stmt

	if statement, err = a.db.Prepare("insert into todo (task) values (?)"); err != nil {
		log.Fatal(err)
	}

	defer statement.Close()

	res, err := statement.Exec("dummy task 1")

	if err != nil {
		log.Fatal(err)

	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)

		log.Printf("ID = %d", lastId)
	}
	return 0, err

	// do something to retrieve the inserted ID from result and return it

}

func (a *app) done(taskID int) error {
	var err error
	var statement *sql.Stmt

	if statement, err = a.db.Prepare("update into todo (task) values(?)"); err != nil { // Update the query here to an update query
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(taskID)

	return err
}

func (a *app) handleGetAllTodos(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.getAll()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (a *app) handleAddTodo(w http.ResponseWriter, r *http.Request) {
	var task NewTodo
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&task); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	id, err := a.addTodo(task.Task)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Todo{id, task.Task, false})
}

func (a *app) handleGetTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Task Id, must be integer")
		return
	}

	todos, err := a.getTodo(id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, todos)
}

func (a *app) handleTodoDone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Task Id, must be integer")
		return
	}

	err = a.done(id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, fmt.Sprintf("marking todo = %d, as done", id))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *app) run() {
	a.server = &http.Server{
		Addr:         ":5000",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.router,
	}

	log.Println("Starting Todo Server on port 5000")
	log.Fatal(a.server.ListenAndServe())
}

func main() {
	a, err := NewApp("./todo.db")

	if err != nil {
		log.Fatal(err)
	}

	a.run()
}

func NewApp(dbPath string) (a *app, err error) {
	a = &app{}

	if err := a.initializeDb("./todo.db"); err != nil {
		return nil, fmt.Errorf("Could not initialize database: %s", err)
	}

	a.initializeRoutes()

	return a, nil
}
