package main

import "net/http"

func main() {
	http.ListenAndServe(":8084", http.FileServer(http.Dir("./docs")))
}
