// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"spotify-go-notion/constant"
	spotify "spotify-go-notion/core"
	notionapi "spotify-go-notion/notion"
	"spotify-go-notion/utils"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	// 在这里修改当前程序访问Spotify的 scope
	auth                                           = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistReadPrivate)
	ch                                             = make(chan *spotify.Client)
	state                                          = "abc123"
	user_id                                        = "justinzwd"
	notion_wuji_integration_secret notionapi.Token = "secret_SUfSLzcHxScCwG88L6yPcvkJZzlSqjKjek3g6457guc"
	notion_wuji_user_id                            = "d0e5814d-d4a3-415a-a7fe-f8c125c19116"
	// notion_wuji_default_database_id notionapi.DatabaseID = "edf16fcf64c645ebb54ece3a48777380"
	notion_wuji_default_database_id notionapi.DatabaseID = "b42b7c458f194f5b840b81e8825e2782"
	track_info_external_spotify                          = "spotify"
)

type TrackDetail struct {
	TrackName  string
	TrackID    string
	TrackCover string
	AlbumName  string
	AlbumID    string
	AlbumCover string
	OpenURL    string
	Artist     []string
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

	// all my liked songs
	// err = extractArtistSongsAndCreateNewPlaylist(client, "7g6jtS5UzZgTtlnuFhhLmT", "Michael Bublé")
	// 2022年度音乐总结
	err = extractArtistSongsAndCreateNewPlaylist(client, "5sQj1V4Z4F5exTR2n0efxj", "Michael Bublé")
	if err != nil {
		log.Fatal(err)
		return
	}

	// fullPlaylistStr, _ := json.Marshal(fullPlaylist)
	// fullPlaylistBytes := []byte(fullPlaylistStr)
	// ioutil.WriteFile("./fullPlaylist.json", fullPlaylistBytes, 0666)
}

func main2() {
	notionClient := notionapi.NewClient(notion_wuji_integration_secret)
	database, err := notionClient.Database.Get(context.Background(), notion_wuji_default_database_id)
	if err != nil {
		log.Fatal(err)
		return
	}
	bb, _ := json.Marshal(database)
	ioutil.WriteFile("./database.json", bb, 0666)
}

func extractArtistSongsAndCreateNewPlaylist(client *spotify.Client, playlistId spotify.ID, artistName string) error {
	//1. 获取播放列表
	playlist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return err
	}

	//2. 获取播放列表所有歌曲以及艺人和歌曲map
	artistSongMap, allArtists, allSongs, err := GetEntirePlaylist(client, playlistId, playlist)

	// artistSongMapStr, _ := json.Marshal(artistSongMap)
	// artistSongMapBytes := []byte(artistSongMapStr)
	// ioutil.WriteFile("./artistSongMap.json", artistSongMapBytes, 0666)
	// return nil

	if songs, ok := artistSongMap[artistName]; ok {
		fmt.Println(artistName, "在此播放列表里一共有", len(songs), "首歌")
		//3. 创建歌单
		newPlaylist, err := client.CreatePlaylistForUser(user_id, artistName, artistName+"'s songs", true)
		if err != nil {
			return err
		}
		fmt.Println("newPlaylist.ID", newPlaylist.ID)

		// 插入多条 tracks 到新建的这个列表
		//spotify:track:7IsdzMn6y2yGKuWOjpVL4l 这个函数会自动帮我们拼接，所以只需要后面的id即可
		//当然也可以改变这个函数的拼接方式
		//4. 插入这些歌曲
		_, err = client.AddTracksToPlaylist(newPlaylist.ID, songs...)
		if err != nil {
			// 非main里面不要用 fatal，保证可以正常将error传递回去
			return err
		}
		fmt.Println("成功插入", artistName, "的歌曲到新播放列表", newPlaylist.Name, "中～")
	}

	//BatchInsertSongsIntoNotion
	//5. 将播放列表所有歌曲插入到notion中
	notionClient := notionapi.NewClient(notion_wuji_integration_secret)
	err = BatchInsertSongsIntoNotion(notionClient, allSongs, allArtists, notion_wuji_default_database_id)
	if err != nil {
		return err
	}

	return nil
}

func GetEntirePlaylist(client *spotify.Client, playlistId spotify.ID, playlist *spotify.FullPlaylist) (map[string][]spotify.ID, []string, []TrackDetail, error) {
	// 1.1 整个播放列表的 totalNum
	totalNum := playlist.Tracks.Total
	fmt.Println("totalNum", totalNum)

	// 1.2 分页 for循环拿数据，append 到数组里面
	count := totalNum / 50
	if totalNum%50 != 0 {
		count++
	}

	// 2. 循环遍历数组，filter，把符合条件的歌放到另外一个数组/map里
	offset := 0
	limit := 50
	// fullTracks := make([]spotify.SimpleTrack, 0)

	// 艺人及其歌曲映射map
	artistSongMap := make(map[string][]spotify.ID, 0)

	// 所有的艺人 array
	allArtistsMap := make(map[string]interface{}, 0)

	// 所有歌曲
	allSongs := make([]TrackDetail, 0)

	for i := 0; i < count; i++ {
		fullPlaylist, err := client.GetPlaylistTracks(playlistId, offset, limit)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, track := range fullPlaylist.Tracks {
			if len(track.Track.Artists) > 0 {
				trackDetail := TrackDetail{}
				trackDetail.TrackID = track.Track.ID.String()
				trackDetail.TrackName = track.Track.Name
				if len(track.Track.Album.Images) > 0 {
					trackDetail.TrackCover = track.Track.Album.Images[0].URL
				}
				trackDetail.AlbumID = track.Track.Album.ID.String()
				trackDetail.AlbumName = track.Track.Album.Name
				trackDetail.OpenURL = track.Track.ExternalURLs[track_info_external_spotify]
				trackDetail.Artist = make([]string, 0)
				for _, artist := range track.Track.Artists {
					// 防止出现name为空的情况
					// 因为已经出现了name为空的情况，但是map的key不能为空
					if artist.Name != "" {
						allArtistsMap[artist.Name] = 0
						trackDetail.Artist = append(trackDetail.Artist, artist.Name)

						if songs, ok := artistSongMap[artist.Name]; ok {
							// track.Track.ID 去重 （基本没必要）
							songs = append(songs, track.Track.ID)
							artistSongMap[artist.Name] = songs
						} else {
							songs := make([]spotify.ID, 0)
							songs = append(songs, track.Track.ID)
							artistSongMap[artist.Name] = songs
						}
					}
				}

				allSongs = append(allSongs, trackDetail)
			}
		}
		offset += 50
	}
	allArtists := utils.ConvertMapKeysToStrArr(allArtistsMap)

	return artistSongMap, allArtists, allSongs, nil
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

func BatchInsertSongsIntoNotion(notionClient *notionapi.Client, allSongs []TrackDetail, allArtists []string, databaseID notionapi.DatabaseID) error {
	// 注意，此部分适合第一次执行，不能反复执行
	//1. patch artists
	err := UpdateDatabaseAllArtists(notionClient, allArtists, databaseID)
	if err != nil {
		return err
	}

	//2.insert songs
	wg := &sync.WaitGroup{}
	for _, songDetail := range allSongs {
		wg.Add(1)
		go func(songDetail TrackDetail) {
			defer wg.Done()
			pageCreateRequest := BuildInsertTrackPageRequest(songDetail.TrackName, songDetail.TrackID, songDetail.TrackCover, songDetail.AlbumName, songDetail.AlbumID, songDetail.OpenURL, songDetail.Artist, databaseID)
			_, err := notionClient.Page.Create(context.Background(), &pageCreateRequest)

			//todo 优化 将错误信息用 channel 传出去
			if err != nil {
				fmt.Println()
			}
		}(songDetail)
	}
	wg.Wait()

	return nil
}

func UpdateDatabaseAllArtists(notionClient *notionapi.Client, allArtists []string, databaseID notionapi.DatabaseID) error {
	prop := make(map[string]notionapi.PropertyConfig, 0)
	options := make([]notionapi.Option, 0)
	colors := []notionapi.Color{"brown", "red", "orange", "yellow", "green", "blue", "purple", "pink", "default", "gray"}
	index := 0
	for _, artistName := range allArtists {
		options = append(options, notionapi.Option{
			// ID:    "",
			Name: artistName,
			// 搞一个color的数组，按顺序赋值
			Color: colors[index%len(colors)],
		})
		index++
	}

	prop[constant.NOTION_DATABASE_COLUMN_ARTISTS] = notionapi.MultiSelectPropertyConfig{
		ID:   "Ly%5DB",
		Type: "multi_select",
		MultiSelect: notionapi.Select{
			Options: options,
		},
	}
	databaseUpdateRequest := notionapi.DatabaseUpdateRequest{
		// Title:      []notionapi.RichText{},
		Properties: prop,
	}
	// bb, _ := json.Marshal(databaseUpdateRequest)
	// fmt.Println(string(bb))

	_, err := notionClient.Database.Update(context.Background(), databaseID, &databaseUpdateRequest)

	if err != nil {
		return err
	}

	return nil
}

func BuildInsertTrackPageRequest(track, trackID, trackCover, album, albumID, openURL string, artists []string, databaseID notionapi.DatabaseID) notionapi.PageCreateRequest {
	pageProp := make(map[string]notionapi.Property, 0)

	//Track name
	trackRichTexts := make([]notionapi.RichText, 0)
	trackRichTexts = append(trackRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: track,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_TRACK] = notionapi.TitleProperty{
		// ID:    "",
		// Type:  "",
		Title: trackRichTexts,
	}

	//TrackID
	trackIDRichTexts := make([]notionapi.RichText, 0)
	trackIDRichTexts = append(trackIDRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: trackID,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_TRACK_ID] = notionapi.RichTextProperty{
		// ID:       "",
		// Type:     "",
		RichText: trackIDRichTexts,
	}

	//Track cover
	trackCoverRichTexts := make([]notionapi.RichText, 0)
	trackCoverRichTexts = append(trackCoverRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: trackCover,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_TRACK_COVER] = notionapi.RichTextProperty{
		// ID:       "",
		// Type:     "",
		RichText: trackCoverRichTexts,
	}

	//Album name
	albumRichTexts := make([]notionapi.RichText, 0)
	albumRichTexts = append(albumRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: album,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_ALBUM] = notionapi.RichTextProperty{
		// ID:       "",
		// Type:     "",
		RichText: albumRichTexts,
	}

	//AlbumID
	albumIDRichTexts := make([]notionapi.RichText, 0)
	albumIDRichTexts = append(albumIDRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: albumID,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_ALBUM_ID] = notionapi.RichTextProperty{
		// ID:       "",
		// Type:     "",
		RichText: albumIDRichTexts,
	}

	//Open url
	openURLRichTexts := make([]notionapi.RichText, 0)
	openURLRichTexts = append(openURLRichTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: openURL,
		},
	})
	pageProp[constant.NOTION_DATABASE_COLUMN_OPEN_URL] = notionapi.RichTextProperty{
		// ID:       "",
		// Type:     "",
		RichText: openURLRichTexts,
	}

	//Artists
	artistOptions := make([]notionapi.Option, 0)
	for _, v := range artists {
		artistOptions = append(artistOptions, notionapi.Option{
			// ID:    "",
			Name: v,
			// Color: "",
		})
	}
	pageProp[constant.NOTION_DATABASE_COLUMN_ARTISTS] = notionapi.MultiSelectProperty{
		ID:          "",
		Type:        "",
		MultiSelect: artistOptions,
	}

	pageCreateRequest := notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: databaseID,
		},
		Properties: pageProp,
		// Children:   []notionapi.Block{},
		// Icon:  &notionapi.Icon{},
		// Cover: &notionapi.Image{},
	}

	return pageCreateRequest
}
