package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/antchfx/htmlquery"
	vkapi "github.com/himidori/golang-vk-api"
)

func main() {

	t := time.Date(2021, time.January, 9, 20, 0, 0, 0, time.Local)
	for i := 0; i < 36; i++ {
		createPost("https://thiscatdoesnotexist.com/", getText(), t)
		t = t.Add(time.Hour * 4)
		fmt.Println(t)
	}

}

func getText() string {

	fileTextUrl := "https://socratify.net/quotes/random"
	doc, err := htmlquery.LoadURL(fileTextUrl)
	if err != nil {
		fmt.Println(err)
	}
	p := htmlquery.FindOne(doc, "//h1[@class='b-quote__text']")

	reqBody, err := json.Marshal(map[string]string{
		"lenght": "30",
		"prompt": p.FirstChild.Data,
	})

	if err != nil {
		print(err)
	}
	resp, err := http.Post("https://pelevin.gpt.dobro.ai/generate/",
		"application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		print(err)
	}
	var res map[string][]string
	json.NewDecoder(resp.Body).Decode(&res)
	return res["replies"][0][1:]
}

func createPost(urlImage string, text string, t time.Time) {
	var vkToken = "9a74a39d513fdf4bb3f06442a8fea21d899c047a220dc8d938f484897cc0534aba3471fff119091e5c1ec"
	var vkGroupID = -199800931

	client, err := vkapi.NewVKClientWithToken(vkToken, &vkapi.TokenOptions{}, true)
	photo, err := client.UploadByLinkGroupWallPhotos(vkGroupID, urlImage)
	if err != nil {
		fmt.Println(err)
	}

	params := url.Values{}
	params.Set("attachments", client.GetPhotosString(photo))
	params.Set("publish_date", strconv.FormatInt(t.Unix(), 10))
	client.WallPost(vkGroupID, text, params)
}
