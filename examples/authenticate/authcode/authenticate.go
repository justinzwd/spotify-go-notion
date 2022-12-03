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
	auth    = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistReadPrivate)
	ch      = make(chan *spotify.Client)
	state   = "abc123"
	user_id = "justinzwd"
)

type Song struct {
	Name string
	ID   string
}

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

	err = extractArtistSongsAndCreateNewPlaylist(client, "7g6jtS5UzZgTtlnuFhhLmT", "Michael Bublé")
	if err != nil {
		log.Fatal(err)
		return
	}

	// fullPlaylistStr, _ := json.Marshal(fullPlaylist)
	// fullPlaylistBytes := []byte(fullPlaylistStr)
	// ioutil.WriteFile("./test2.json", fullPlaylistBytes, 0666)
}

func extractArtistSongsAndCreateNewPlaylist(client *spotify.Client, playlistId spotify.ID, artistName string) error {
	//1. 获取整个播放列表所有的歌曲
	playlist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return err
	}

	//1.1 整个播放列表的 totalNum
	totalNum := playlist.Tracks.Total
	fmt.Println("totalNum", totalNum)

	//1.2 分页 for循环拿数据，append 到数组里面
	count := totalNum / 50
	if totalNum%50 != 0 {
		count++
	}

	//2. 循环遍历数组，filter，把符合条件的歌放到另外一个数组/map里
	offset := 0
	limit := 50
	// fullTracks := make([]spotify.SimpleTrack, 0)
	artistSongMap := make(map[string][]spotify.ID, 0)
	for i := 0; i < count; i++ {
		fullPlaylist, err := client.GetPlaylistTracks(playlistId, offset, limit)
		if err != nil {
			log.Fatal(err)
			return err
		}
		for _, track := range fullPlaylist.Tracks {
			if len(track.Track.Artists) > 0 {
				for _, artist := range track.Track.Artists {
					if songs, ok := artistSongMap[artist.Name]; ok {
						//todo ID需不需要去重呢？
						songs = append(songs, track.Track.ID)
						artistSongMap[artist.Name] = songs
					} else {
						songs := make([]spotify.ID, 0)
						songs = append(songs, track.Track.ID)
						artistSongMap[artist.Name] = songs
					}
				}
			}
		}
		offset += 50
	}

	if songs, ok := artistSongMap[artistName]; ok {
		fmt.Println(artistName, "在此播放列表里一共有", len(songs), "首歌")
		//3. 创建歌单
		newPlaylist, err := client.CreatePlaylistForUser(user_id, artistName, artistName+"'s songs", true)
		if err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Println("newPlaylist.ID", newPlaylist.ID)

		// 插入多条 tracks 到新建的这个列表
		//spotify:track:7IsdzMn6y2yGKuWOjpVL4l 这个函数会自动帮我们拼接，所以只需要后面的id即可
		//当然也可以改变这个函数的拼接方式
		//4. 插入这些歌曲
		_, err = client.AddTracksToPlaylist(newPlaylist.ID, songs...)
		if err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Println("成功插入", artistName, "的歌曲到新播放列表", newPlaylist.Name, "中～")
	}

	return nil
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
