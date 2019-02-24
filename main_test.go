package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
)

var testApp *app

func deleteDb(path string) error {
	// check if the file exists before we attempt to delete it
	if exists, err := fileExists(path); err != nil {
		return err
	} else if !exists {

		return errors.New("file does not exist")
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		// the file may or may not exists, but we do not have access to get file stats
		if os.IsNotExist(err) { // check the error to see if it's because the file doesn't exist
			// the file doesn't exist so we can return a false flag with no error
			return false, nil
		}
		// the file exists but we don't have permission to access it or there
		// is some other error associated with the fil
		return false, err
	}

	// if we get to here, then it's because the file exists and we have permission to access it
	return true, nil
}

func populateTestDb() error {
	var statement *sql.Stmt
	var err error

	if statement, err = testApp.db.Prepare("insert into todo (task) values (?)"); err != nil {
		return err
	}

	defer statement.Close()

	tasks := []string{
		"Dummy Task 1",
		"Dummy Task 2",
		"Dummy Task 3",
	}

	for i, t := range tasks {
		if _, err = statement.Exec(t); err != nil {
			return fmt.Errorf("error occurred whilst inserting task %d - %v", i, err)
		}
	}

	return nil
}

func TestMain(m *testing.M) {
	// create

	dbPath := "./todo_test.db"

	var err error

	testApp, err = NewApp(dbPath)

	if err != nil {
		log.Fatal("Could not create test todo database")
	}

	if err := populateTestDb(); err != nil {
		log.Fatal("Could not create test data for tests")
	}

	flag.Parse()

	//code := m.Run()

	deleteDb(dbPath)

	os.Exit(0)
}

func Test_GetAllTasks(t *testing.T) {
	tasks, err := testApp.getAll()

	if err != nil {
		t.Fatalf("Could not get all tasks - %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	for i, task := range tasks {
		taskID := i + 3
		if task.ID != taskID {
			t.Errorf("Task %d expected id: %d, got %d", taskID, taskID, task.ID)
		}

		if task.Task != fmt.Sprintf("Dummy Task %d", taskID) {
			t.Errorf("Task %d, expected Task Description %s, got %s", taskID, fmt.Sprintf("Dummy Task %d", taskID), task.Task)
		}

		if task.Done != false {
			t.Errorf("Task %d, expected Done = %v, got %v", taskID, false, task.Done)
		}
	}
}

func Test_GetTask(t *testing.T) {
	tasks, err := testApp.getTodo(1)

	if err != nil {
		t.Fatalf("Could not retrieve task - %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 tasks, got %d", len(tasks))
	}

	for i, task := range tasks {
		taskID := i + 1
		if task.ID != taskID {
			t.Errorf("Task %d expected id: %d, got %d", taskID, taskID, task.ID)
		}
	}
}

func Test_AddTask(t *testing.T) {
	var taskID int
	var err error

	if taskID, err = testApp.addTodo("test task 1"); err != nil {
		t.Errorf("could not add task - %v", err)
	}

	if taskID != 4 {
		t.Errorf("taskID was not correct, expected 4, got %d", taskID)
	}
}

func Test_TaskDone(t *testing.T) {
	err := testApp.done(1)

	if err != nil {
		t.Errorf("could not mark done - %v", err)
	}

	tasks, err := testApp.getTodo(1)

	if err != nil {
		t.Errorf("could not retrieve tasks after mark done - %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("%d tasks returned after mark done, expecting 1", len(tasks))
	}

	todo := tasks[0]

	if !todo.Done {
		t.Error("todo not done, after marked done")
	}
}
