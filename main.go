package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-pentview/services"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// func getEnvVar(key string) string {
//    err := godotenv.Load(".env")

//    if err != nil {
//        log.Fatal(err)
//    }
//    return os.Getenv(key)
// }

func main() {
	srv := initServer()
	err := http.ListenAndServe(":3000", srv)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

const dbpath = "data/store.db3"

type Server struct {
	*mux.Router
	repo *services.SQLiteRepository
}

func initServer() *Server {
	os.Remove(dbpath)

	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal(err)
	}

	repo := services.NewSQLiteRepository(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}

	s := &Server{
		Router: mux.NewRouter(),
		repo:   repo,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.HandleFunc("/employee-service/role", s.createRole(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/role", s.getRoles(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user", s.createUser(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/user/list", s.getUsers(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user/{id}", s.updateUser(s.repo)).Methods("PUT")
	s.HandleFunc("/employee-service/user/{id}", s.deleteUser(s.repo)).Methods("DELETE")
	s.HandleFunc("/employee-service/hour-register", s.createClocking(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/hour-register", s.getClockings(s.repo)).Methods("GET")
}

func (s *Server) createRole(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Retrieve json
		var role services.Role
		json.NewDecoder(r.Body).Decode(&role)

		// Create role
		_, err := repo.CreateRole(role)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read role
		roleCreated, _ := repo.GetRoleByName(role.Name)

		// Response role
		data := struct {
			Message string        `json:"message"`
			Role    services.Role `json:"role"`
		}{"Rol creado correctamente", *roleCreated}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getRoles(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		roles, _ := repo.AllRoles()
		if err := json.NewEncoder(w).Encode(roles); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createUser(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Retrieve json
		var user services.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Create user
		_, err = repo.CreateUser(user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read user
		userCreated, _ := repo.GetUserByName(user.Name)

		// Response user
		data := struct {
			Message string        `json:"message"`
			User    services.User `json:"user"`
		}{"Usuario creado", *userCreated}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getUsers(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		users, _ := repo.AllUsers()
		if len(users) == 0 {
			users = []services.User{}
		}
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) updateUser(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Retireve id
		id := mux.Vars(r)["id"]
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		// Retrieve body
		var user services.User
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Create user
		_, err = repo.UpdateUser(intid, user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read user
		userUpdated, _ := repo.GetUserByName(user.Name)

		// Response user
		data := struct {
			Message string        `json:"message"`
			User    services.User `json:"user"`
		}{"Usuario actualizado correctamente", *userUpdated}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) deleteUser(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		err = repo.DeleteUser(intid)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		data := struct {
			Message string `json:"message"`
		}{"Delete user"}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createClocking(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Retrieve json
		var clocking services.Clocking
		err := json.NewDecoder(r.Body).Decode(&clocking)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Create clocking
		res, err := repo.CreateClocking(clocking)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read clocking
		clockingCreated, _ := repo.GetClockingById(res.ClockingID)

		// Response clocking
		data := struct {
			Message  string            `json:"message"`
			Clocking services.Clocking `json:"clocking"`
		}{"Clocking registrado", *clockingCreated}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getClockings(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		clockings, _ := repo.AllClockings()
		if len(clockings) == 0 {
			clockings = []services.Clocking{}
		}
		if err := json.NewEncoder(w).Encode(clockings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
