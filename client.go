package main

import (
	"github.com/shpaktakur1/TestAvito/db"
	"github.com/shpaktakur1/TestAvito/RestAPI"
	"math/rand"
	"os"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getRandomStrings(n int, l int) []string {
	result := make([]string, n)
	for i := 0; i < n; i++ {
		b := make([]byte, l)
		for j := range b {
			b[j] = letters[rand.Intn(len(letters))]
		}
		result[i] = string(b)
	}
	return result
}

func startApps() {
	app1 := rest.App{}
	app1.Authorization = &rest.BasicAuthorizer{"User1", "Password1"}
	err := app1.Initialize(0, os.Stdout, nil, 500, 1, nil)
	if err != nil {
		panic(err)
	}
	go app1.Run(":8080", 10, 10)

	app2 := rest.App{}
	err = app2.Initialize(0, os.Stderr, nil, 500, 1, nil)
	if err != nil {
		panic(err)
	}
	go app2.Run(":8081", 10, 10)
}

func main() {
	startApps()
	c1 := rest.NewClient("http://localhost:8080", 10, "User1", "Password1")
	c2 := rest.NewClient("http://localhost:8081", 10, "", "")

	c, err := db.Shard(nil, false, c1, c2)
	if err != nil {
		panic(err)
	}

	for i, str := range getRandomStrings(1200, 8) {
		c.Set(str, []interface{}{i, str}, 0)
		c.Get(str)
		c.Remove(str)
	}
}
