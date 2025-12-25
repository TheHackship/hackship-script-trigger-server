package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"log"
)

func responseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "POST request received successfully!")
}

func jsonParser(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "fail"
	}

	var config map[string]map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		return "jsonParser: error"
	}

	return "jsonParser: success"
}

func main() {
	rootDir, _ := os.Getwd()
	filePath := rootDir + "/config.json"

	result := jsonParser(filePath)
	fmt.Println(result)

	http.HandleFunc("/", responseHandler);
	fmt.Println("Server running on http://localhost:8080")
	err := http.ListenAndServe("8080", nil)
	
	if err != nil{
		log.Fatal("HTTP Server Error: ", err)
	}
}
