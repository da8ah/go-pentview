package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-pentview/services"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
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

func main() {
	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Accept", "Accept-Language", "Content-Type", "Content-Language", "Content-Disposition", "Origin"})
	originsOk := handlers.AllowedOrigins([]string{getEnvVar("CORS")})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})

	srv := initServer()
	err := http.ListenAndServe(":"+getEnvVar("PORT"), handlers.CORS(originsOk, headersOk, methodsOk)(srv.Router))
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

	user := services.UserToCreate{UserID: 1, Name: "Admin", Last: "Admin", Email: "admin@yopmail.com", Password: "admin@2024", PFP: "upload/nopfp.png", CreatedAt: "today", Role: "1"}
	s.repo.CreateUser(user)

	clocking := services.Clocking{Type: "in", Date: time.Now().Format(time.RFC3339), UserID: 1}
	s.repo.CreateClocking(clocking)
}

func (s *Server) routes() {
	s.HandleFunc("/upload/{img}", s.getPFP()).Methods("GET")
	s.HandleFunc("/employee-service/user/auth/login", s.login(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/user/profile", s.getProfile(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user/update-profile", s.updateProfile(s.repo)).Methods("PUT")
	s.HandleFunc("/employee-service/role", s.createRole(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/role", s.getRoles(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user", s.createUser(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/user/list", s.getUsers(s.repo)).Methods("GET")
	s.HandleFunc("/employee-service/user/{id}", s.updateUser(s.repo)).Methods("PUT")
	s.HandleFunc("/employee-service/user/{id}", s.deleteUser(s.repo)).Methods("DELETE")
	s.HandleFunc("/employee-service/hour-register", s.createClocking(s.repo)).Methods("POST")
	s.HandleFunc("/employee-service/hour-register", s.getClockings(s.repo)).Methods("GET")
}

func generarToken(user_id int64, username string, role string) string {
	key := []byte(getEnvVar("TOKEN"))
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username":  username,
			"sub":       user_id,
			"authority": role,
			"iat":       time.Now().UnixMilli(),
			"exp":       time.Now().Add(time.Hour * 1).UnixMilli(),
		})
	s, _ := t.SignedString(key)
	return s
}
func validateToken(token string) (bool, int64) {
	tokenDecoded, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to parse")
		}
		return []byte(getEnvVar("TOKEN")), nil
	})
	if err != nil {
		return false, 0
	}
	if claims, ok := tokenDecoded.Claims.(jwt.MapClaims); ok && tokenDecoded.Valid {
		sub := fmt.Sprint(claims["sub"])
		if len(sub) > 0 {
			id, err := strconv.ParseInt(sub, 10, 64)
			if err != nil {
				panic(err)
			}
			return tokenDecoded.Valid, id
		}
	}
	return tokenDecoded.Valid, 0
}

func UploadPFP(r *http.Request) {
	file, handler, err := r.FormFile("image")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	f, err := os.OpenFile("data/img/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, _ = io.Copy(f, file)
	fmt.Printf("File %s saved!\n", handler.Filename)
}
func (s *Server) getPFP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imgpath := mux.Vars(r)["img"]
		img, err := os.Open("data/img/" + imgpath)
		if err != nil {
			log.Fatal(err) // perhaps handle this nicer
		}
		defer img.Close()
		w.Header().Set("Content-Type", "image/png") // <-- set the content-type header
		io.Copy(w, img)
	}
}

func (s *Server) login(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Retrieve json
		var credentials services.Credentials
		json.NewDecoder(r.Body).Decode(&credentials)

		user, err := repo.CompareCredentials(credentials)
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
		}{generarToken(user.UserID, user.Email, user.Role.Name)}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) getProfile(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
			user_id int64
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid, user_id = validateToken(token)
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

		profile, _ := repo.GetProfileById(user_id)
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func (s *Server) updateProfile(repo *services.SQLiteRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth
		auth := r.Header.Get("Authorization")
		var (
			token   string
			isValid bool
			user_id int64
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid, user_id = validateToken(token)
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

		// Retrieve body
		var body services.PutProfile
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Update profile
		profile, _ := repo.GetProfileById(user_id)
		profile.Name = body.Name
		profile.Last = body.Last
		profile.Email = body.Email
		_, err = repo.UpdateProfile(user_id, *profile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Response profile
		res := struct {
			Message string        `json:"message"`
			User    services.User `json:"user"`
		}{"Usuario actualizado correctamente", *profile}

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
			isValid, _ = validateToken(token)
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
		roleCreated, err := repo.CreateRole(role)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
			isValid, _ = validateToken(token)
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
		data := struct {
			Data []services.Role `json:"data"`
		}{Data: roles}
		if err := json.NewEncoder(w).Encode(data); err != nil {
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
			isValid, _ = validateToken(token)
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

		// Retrieve files
		UploadPFP(r)
		body := r.FormValue("json")

		// Parse json
		var userToCreate services.UserToCreate
		err := json.Unmarshal([]byte(body), &userToCreate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}
		userToCreate.PFP = "upload/" + userToCreate.PFP

		// Create user
		err = repo.CreateUser(userToCreate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Response user
		res := struct {
			Message string `json:"message"`
		}{"Usuario creado"}
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
			isValid, _ = validateToken(token)
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
			isValid, _ = validateToken(token)
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
		var userToUpdate services.User
		err = json.NewDecoder(r.Body).Decode(&userToUpdate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Create user
		user, err := repo.UpdateUser(intid, userToUpdate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Read user
		userUpdated, _ := repo.GetUserById(user.UserID)

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
			isValid, _ = validateToken(token)
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
			user_id int64
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid, user_id = validateToken(token)
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
		clocking.UserID = user_id

		// Create clocking
		clockingCreated, err := repo.CreateClocking(clocking)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := struct {
				Message string `json:"message"`
			}{Message: err.Error()}
			json.NewEncoder(w).Encode(msg)
			return
		}

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
			user_id int64
		)
		if len(auth) > 0 {
			token = strings.Split(auth, " ")[1]
			isValid, user_id = validateToken(token)
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

		clockings, _ := repo.AllClockings(user_id)
		if len(clockings) == 0 {
			clockings = []services.Clocking{}
		}
		data := struct {
			Data []services.Clocking `json:"data"`
		}{Data: clockings}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
