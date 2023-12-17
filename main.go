package main

import (
	_ "github.com/lib/pq"
	"github.com/m0rk0vka/go-test/router"
)


func main() {
	app := router.Router()
	app.Listen(":8080")
}
