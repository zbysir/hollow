package main

import "net/http"

func main() {
	// http://localhost:8084/blog
	http.ListenAndServe(":8084", http.StripPrefix("/blog", http.FileServer(http.Dir("./docs"))))
}
