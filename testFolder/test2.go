package main

import (
	"encoding/base64"
	"fmt"
)

func main1() {
	clientId := "88f46cd5cc784dd897218e862d39b366"
	clientSecret := "bfc246da5a7045a6a2390a6b59987a9d"

	clientBytes := []byte(clientId + ":" + clientSecret)
	sEnc := base64.StdEncoding.EncodeToString(clientBytes)
	fmt.Printf("%s\n", sEnc)
}
