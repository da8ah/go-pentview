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
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func getEnvVar(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}
	return os.Getenv(key)
}

func generarToken(user_id int64, username string, role string) string {
	key := []byte(getEnvVar("TOKEN"))
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username":  username,
			"sub":       user_id,
			"authority": role,
			"iat":       strconv.FormatInt(time.Now().UnixMilli(), 10),
			"exp":       strconv.FormatInt(time.Now().Add(time.Hour*1).UnixMilli(), 10),
		})
	s, _ := t.SignedString(key)
	return s
}

func main() {
	srv := initServer()
	err := http.ListenAndServe(":"+getEnvVar("PORT"), srv)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

type Server struct {
	*mux.Router
	repo *services.SQLiteRepository
}

func initServer() *Server {
	os.Remove(getEnvVar("DBPATH"))

	db, err := sql.Open("sqlite3", getEnvVar("DBPATH"))
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
	s.createAdminUser()
	s.routes()
	return s
}

func (s *Server) createAdminUser() {
	role := services.Role{RoleID: 1, Name: "admin"}
	s.repo.CreateRole(role)

	user := services.User{UserID: 1, Name: "Admin Admin", Email: "admin@yopmail.com", Password: "Admin2022", PFP: "nopfp", CreatedAt: "today", RoleID: 1}
	s.repo.CreateUser(user)
}

func (s *Server) routes() {
	s.HandleFunc("/employee-service/user/auth/login", s.login(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/role", s.createRole(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/role", s.getRoles(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user", s.createUser(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/user/list", s.getUsers(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user/{id}", s.updateUser(s.repo)).Methods("PUT")
	s.HandleFunc("/employee-service/user/{id}", s.deleteUser(s.repo)).Methods("DELETE")
	s.HandleFunc("/employee-service/hour-register", s.createClocking(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/hour-register", s.getClockings(s.repo)).Methods("GET")
}

func validateToken(token string) bool {
	res, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to parse")
		}
		return getEnvVar("Token"), nil
	})
	return res != nil
}
func (s *Server) login(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Retrieve json
		var credentials services.Credentials
		json.NewDecoder(r.Body).Decode(&credentials)

		user, role, err := repo.CompareCredentials(credentials)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		res := struct {
			Token string `json:"access_token"`
		}{generarToken(user.UserID, user.Email, role.Name)}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createRole(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
		res := struct {
			Message string        `json:"message"`
			Role    services.Role `json:"role"`
		}{"Rol creado correctamente", *roleCreated}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getRoles(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
		res := struct {
			Message string        `json:"message"`
			User    services.User `json:"user"`
		}{"Usuario creado", *userCreated}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getUsers(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
		res := struct {
			Message string        `json:"message"`
			User    services.User `json:"user"`
		}{"Usuario actualizado correctamente", *userUpdated}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) deleteUser(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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

		res := struct {
			Message string `json:"message"`
		}{"Delete user"}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createClocking(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
		resCreated, err := repo.CreateClocking(clocking)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read clocking
		clockingCreated, _ := repo.GetClockingById(resCreated.ClockingID)

		// Response clocking
		res := struct {
			Message  string            `json:"message"`
			Clocking services.Clocking `json:"clocking"`
		}{"Clocking registrado", *clockingCreated}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) getClockings(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid = validateToken(token)
		} else {
			isValid = false
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			msg := struct {
				Message string `json:"message"`
			}{Message: "no autorizado"}
			json.NewEncoder(w).Encode(msg)
			return
		}

		var user_id_from_token int64 = 1
		clockings, _ := repo.AllClockings(user_id_from_token)
		if len(clockings) == 0 {
			clockings = []services.Clocking{}
		}
		if err := json.NewEncoder(w).Encode(clockings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
