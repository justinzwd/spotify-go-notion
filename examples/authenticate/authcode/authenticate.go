// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//       - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"fmt"
	"log"
	"net/http"

	spotify "spotify-go-notion/core"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	//todo 在这里修改 scope
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	fullPlaylist, err := client.CreatePlaylistForUser(user.ID, "new playlist1234", "this is just a description", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fullPlaylist.ID)

	//todo 插入多条 tracks 到新建的这个列表
	//spotify:track:7IsdzMn6y2yGKuWOjpVL4l 这个函数会自动帮我们拼接，所以只需要后面的id即可
	//当然也可以改变这个函数的拼接方式
	_, err = client.AddTracksToPlaylist(fullPlaylist.ID, "7IsdzMn6y2yGKuWOjpVL4l", "3mWZefa0yfuRz0KjeeVIBU")
	if err != nil {
		log.Fatal(err)
	}

	// todo client.GetPlaylist 获取到 totalNum
	// 然后再每50去请求，将数据全部保存在本地
	// 也可以指定字段，只获取自己想要的字段

	// fullPlaylist, err := client.GetPlaylistTracks("7g6jtS5UzZgTtlnuFhhLmT", 100, 100)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fullPlaylistStr, _ := json.Marshal(fullPlaylist)
	// fullPlaylistBytes := []byte(fullPlaylistStr)
	// ioutil.WriteFile("./test2.json", fullPlaylistBytes, 0666)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}
