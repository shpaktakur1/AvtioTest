package rest

import (
	"github.com/gorilla/mux"
	"github.com/shpaktakur1/TestAvito/db"
	"io"
	"log"
	"net/http"
	"time"
)

type App struct {
	initialized bool

	Authorization Authorizer
	Router        *mux.Router
	Cache         db.Cache
}

// Начало проверки по представленному адресу
func (a *App) Run(addr string, readTimeout int, writeTimeout int) {
	if readTimeout == 0 {
		readTimeout = 15
	}
	if writeTimeout == 0 {
		writeTimeout = 15
	}

	if !a.initialized {
		log.Fatal("Canot run uninitialized application. Execute Initialize() method first")
	}

	srv := &http.Server{
		Handler:      a.Router,
		Addr:         addr,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}

// Инициализация кэша и маршрутизации
func (a *App) Initialize(defaultTTL time.Duration, out io.Writer, rw io.ReadWriter, saveFreq time.Duration, nShards int, shardFunction func(string) uint32) (err error) {
	a.Cache, err = db.NewCache(
		defaultTTL,
		out,
		rw,
		saveFreq,
		nShards,
		shardFunction,
	)
	if err != nil {
		return err
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
	a.initialized = true
	return nil
}



func (a *App) initializeRoutes() {
	wrappers := []wrapper{}
	if a.Authorization != nil {
		wrappers = append(wrappers, auth(a.Authorization))
	}
	a.Router.HandleFunc("/{key}/{index}", Wrap(a.actionGetByIndex, wrappers)).Methods("GET")
	a.Router.HandleFunc("/{key}", Wrap(a.actionGet, wrappers)).Methods("GET")
	a.Router.HandleFunc("/{key}", Wrap(a.actionSet, wrappers)).Methods("POST")
	a.Router.HandleFunc("/{key}", Wrap(a.actionRemove, wrappers)).Methods("DELETE")
	a.Router.HandleFunc("/", Wrap(a.actionKeys, wrappers)).Methods("GET")
}

func (a *App) actionGetByIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	result, err := a.Cache.GetAtIndex(vars["key"], vars["index"])
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, result)
}

func (a *App) actionGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	value, err := a.Cache.Get(vars["key"])
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, value)
}

func (a *App) actionSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	ttl, err := processTTL(q.Get("ttl"))

	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}

	t, err := decodeJSONBody(r.Body)
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	value, err := a.Cache.Set(vars["key"], t, ttl)
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, value)
}

func (a *App) actionRemove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := a.Cache.Remove(vars["key"])
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, "OK")
}

func (a *App) actionKeys(w http.ResponseWriter, r *http.Request) {
	result, err := a.Cache.Keys()
	if err != nil {
		respondWithAppError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, result)
}
