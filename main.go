package main

import (
	"database/sql"
	"errors"
	"fmt"
	"go-pentview/services"
	"io"
	"log"
	"net/http"
	"os"
)

const fileName = "data/store.db3"

/*
func createExamples(repository *services.SQLiteRepository) {
	gosamples := services.User{
		Name:  "Alejandro Serrano",
		Email: "aserrano@yopmail.com",
		Role:  1,
	}
	golang := services.User{
		Name:  "Carolina Chamba",
		Email: "cchamba@yopmail.com",
		Role:  2,
	}

	createdGosamples, err := repository.Create(gosamples)
	if err != nil {
		log.Fatal(err)
	}
	createdGolang, err := repository.Create(golang)
	if err != nil {
		log.Fatal(err)
	}
	gotGosamples, err := repository.GetByName("Alejandro Serrano")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("get by name: %+v\n", gotGosamples)

	all, err := repository.All()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAll users:\n")
	for _, user := range all {
		fmt.Printf("user: %+v\n", user)
	}

	createdGosamples.Role = 1
	if _, err := repository.Update(createdGosamples.ID, *createdGosamples); err != nil {
		log.Fatal(err)
	}

	if err := repository.Delete(createdGolang.ID); err != nil {
		log.Fatal(err)
	}

	all, err = repository.All()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nAll users:\n")
	for _, user := range all {
		fmt.Printf("user: %+v\n", user)
	}
}
*/

func main() {
	os.Remove(fileName)

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		log.Fatal(err)
	}

	repository := services.NewSQLiteRepository(db)
	if err := repository.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Routes
	http.HandleFunc("/", getHome)
	http.HandleFunc("/employee-service/user/list", getUsers)

	err = http.ListenAndServe(":3000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getHome(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World!")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World!")
}
