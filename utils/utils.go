package utils

import (
	"encoding/json"
	"log"
	"regexp"
)

func IsValidEmail(email string) bool {
	reg, _ := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return reg.MatchString(email)
}

func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func JsonEncode(data interface{}) string {
	d, _ := json.Marshal(data)
	return string(d)
}
