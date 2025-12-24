package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func jsonParser(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "fail"
	}

	var config map[string]map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		return "jsonParser: failed"
	}

	return "jsonParser: success"
}

func main() {
	rootDir, _ := os.Getwd()
	filePath := rootDir + "/config.json"

	result := jsonParser(filePath)
	fmt.Println(result)
}
