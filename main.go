package main

import (
	"flag"
	"github.com/shpaktakur1/TestAvito/rest"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "http server address")
	readTimeout := flag.Int("readTimeout", 10, "http read timeout")
	writeTimeout := flag.Int("writeTimeout", 10, "http write timeout")

	defaultTtl := flag.Int("defaultTTL", 0, "default ttl in seconds for every entry")
	nShards := flag.Int("shards", 1, "number of shards for concurrent writes")

	login := flag.String("login", "", "login for basic auth")
	password := flag.String("password", "", "password for basic auth")

	filename := flag.String("file", "", "database path")
	saveFreq := flag.Int("saveFreq", 500, "save to disk frequency in ms")

	logTo := flag.String("log", "", "stdout/stderr/path_to_log_file. Does not log if empty")

	flag.Parse()
	app := rest.App{}

	if *login != "" && *password != "" {
		app.Authorization = &rest.BasicAuthorizer{Username: *login, Password: *password}
	}

	var rw io.ReadWriteCloser
	var err error

	if *filename != "" {
		rw, err = os.OpenFile(*filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer rw.Close()
	}

	var writer io.Writer = nil
	if *logTo != "" {
		switch *logTo {
		case "stdout":
			writer = os.Stdout
		case "stderr":
			writer = os.Stderr
		default:
			f, err := os.OpenFile(*logTo, os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			writer = f
		}
	}

	err = app.Initialize(
		time.Duration(*defaultTtl)*time.Second,
		writer,
		rw,
		time.Duration(*saveFreq)*time.Millisecond,
		*nShards,
		nil)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(*addr, *readTimeout, *writeTimeout)
}
