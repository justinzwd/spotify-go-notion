package ClientCredentialFlow

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main2() {

	url := "https://api.spotify.com/v1/users/justinzwd/playlists"
	method := "POST"

	payload := strings.NewReader(`{
    "name": "New12 Playlist",
    "description": "New playlist description",
    "public": true
	}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Bearer BQDFgD6ObKa9chda-WRQjrP6e-pxGid3MlSCzD4nxjbQ5J6ETUHWnmX4dV3WYDcLjrVTLrszETvQU7wYE30GrHs21fFwiLHBt4kjzVuPIJjj0KCgrA8shTGpNrdrxkI63xZ3lj24JdZY--4fgjeGzdMyKIqGVJ_dLcOyY26qf1D3lPQbAbTJIL0tjN6rfB4uevFGQPbg1eCki23Buj93I63wJmZXmibTRWfD38tajEnA5sGzNBsi")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
