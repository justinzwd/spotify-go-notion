package ClientCredentialFlow

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"spotify-go-notion/core"
	"strings"
)

const (
	CLIENT_ID           = "88f46cd5cc784dd897218e862d39b366"
	CLIENT_SECRET       = "bfc246da5a7045a6a2390a6b59987a9d"
	USER_ID             = "justinzwd"
	DEFAULT_PLAYLIST_ID = "7g6jtS5UzZgTtlnuFhhLmT"
)

func main() {
	// fmt.Println(os.Getenv("SPOTIFY_ID"))
	// fmt.Println(os.Getenv("SPOTIFY_SECRET"))
}

func createPlaylist(userID, playlistName, description string, isPublic bool) error {
	url := "https://api.spotify.com/v1/users/" + userID + "/playlists"
	method := "POST"

	requestStr := `{
		"name": "` + playlistName + `",
		"description": "` + description + `",
		"public": ` + fmt.Sprintf("%v", isPublic) + `
	}`
	fmt.Println("requestStr", requestStr)
	payload := strings.NewReader(requestStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	//todo 把token缓存起来
	accessToken := getAuthToken(CLIENT_ID, CLIENT_SECRET)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))
	return nil
}

func getPlaylist(playlistId string) (core.FullPlaylist, error) {
	playlist := core.FullPlaylist{}
	url := "https://api.spotify.com/v1/playlists/" + playlistId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		// fmt.Println(err)
		return playlist, err
	}
	accessToken := getAuthToken(CLIENT_ID, CLIENT_SECRET)
	// fmt.Println("accessToken", accessToken)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		// fmt.Println(err)
		return playlist, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// fmt.Println(err)
		return playlist, err
	}
	// fmt.Println(string(body))
	// ioutil.WriteFile("./getPlaylist.txt", body, 0666)
	err = json.Unmarshal(body, &playlist)
	if err != nil {
		// fmt.Println(err)
		return playlist, err
	}
	return playlist, nil
}

func getAuthToken(clientId, clientSecret string) string {
	clientBytes := []byte(clientId + ":" + clientSecret)
	base64Str := base64.StdEncoding.EncodeToString(clientBytes)

	url := "https://accounts.spotify.com/api/token"
	method := "POST"

	payload := strings.NewReader("grant_type=client_credentials")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Authorization", "Basic "+base64Str)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "__Host-device_id=AQBcKZYS2kmjxmuF-2wBIvObC9Z0XK6U6LrnbPwR-K5XQi03ZPhYOTXXo9LBFLWBFL7R4-e9RQPJ97e6tDQMW7y1kUkZO1BVZT0; sp_tr=false")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	authResp := AuthGetResponse{
		AccessToken: "",
		TokenType:   "",
		ExpiresIn:   0,
	}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return authResp.AccessToken
}
