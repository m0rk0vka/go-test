package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/m0rk0vka/go-test/router"
)


func main() {
	app := router.Router()
	app.Listen(os.Getenv("addr"))
}
