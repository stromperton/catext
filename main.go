package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	vkapi "github.com/himidori/golang-vk-api"
	tb "gopkg.in/tucnak/telebot.v2"
)

const AdminID int = 303629013
const vkGroupID = -199800931

var vkToken = ""

var (
	RBtnCreatePosts = tb.ReplyButton{Text: "Заготовить посты"}

	IBtnCreate   = tb.InlineButton{Text: "✔️ Опубликовать", Unique: "ok"}
	IBtnEditText = tb.InlineButton{Text: "✏️ Редактировать текст", Unique: "editText"}
	IBtnAddText  = tb.InlineButton{Text: "🔗 Дополнить текст", Unique: "addText"}
	IBtnReText   = tb.InlineButton{Text: "🔄 Обновить текст", Unique: "reText"}
	IBtnReCat    = tb.InlineButton{Text: "🔄 Обновить кота", Unique: "reCat"}
	InlinePost   = &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{{IBtnCreate}, {IBtnEditText, IBtnAddText}, {IBtnReText, IBtnReCat}},
	}
)

func main() {

	var (
		port      = os.Getenv("PORT")
		publicURL = os.Getenv("PUBLIC_URL")
		token     = os.Getenv("TOKEN")

		SuperTimer         time.Time
		IsBotStateEditText = false
	)
	vkToken = os.Getenv("TOKEN_VK")

	poller := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	middle := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		if upd.Message != nil {
			if upd.Message.Sender.ID == AdminID {
				return true
			}
		} else {
			if upd.Callback.Sender.ID == AdminID {
				return true
			}
		}
		return false
	})

	b, err := tb.NewBot(tb.Settings{
		Token:     token,
		Poller:    middle,
		ParseMode: tb.ModeHTML,
	})

	if err != nil {
		fmt.Println(err)
	}

	b.Handle("/start", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		DownloadFile("cat.jpg", "https://thiscatdoesnotexist.com/")
		text := getText("CITATA")
		_, err := b.Send(m.Sender, &tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)

		SuperTimer = getLastPostTimeVK()
		fmt.Println(SuperTimer, err)
	})

	b.Handle(&IBtnCreate, func(c *tb.Callback) {
		b.Respond(c)
		SuperTimer = SuperTimer.Add(time.Hour * 4)
		fmt.Println(c.Message.Caption)
		createPostVK("cat.jpg", c.Message.Caption, SuperTimer)

		DownloadFile("cat.jpg", "https://thiscatdoesnotexist.com/")
		text := getText("CITATA")
		_, err := b.Edit(c.Message, &tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)

		SuperTimer = getLastPostTimeVK()
		fmt.Println(SuperTimer, err)
	})
	b.Handle(&IBtnReText, func(c *tb.Callback) {
		b.Respond(c)
		text := getText("CITATA")
		b.Edit(c.Message, &tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)
	})
	b.Handle(&IBtnReCat, func(c *tb.Callback) {
		b.Respond(c)
		DownloadFile("cat.jpg", "https://thiscatdoesnotexist.com/")
		b.Edit(c.Message, &tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: c.Message.Caption,
		}, InlinePost)
	})
	b.Handle(&IBtnEditText, func(c *tb.Callback) {
		b.Respond(c)
		b.Send(c.Sender, "Отправь новый текст")
		IsBotStateEditText = true
	})

	b.Handle(&IBtnAddText, func(c *tb.Callback) {
		b.Respond(c)
		text := getText(c.Message.Caption)
		b.Edit(c.Message, &tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if IsBotStateEditText {
			IsBotStateEditText = false
			b.Send(m.Sender, &tb.Photo{
				File:    tb.FromDisk("cat.jpg"),
				Caption: m.Text,
			}, InlinePost)
		}
	})

	b.Start()

}

/*
func CreatePosts() {

	t := time.Date(2021, time.January, 21, 20, 0, 0, 0, time.Local)
	for i := 0; i < 72; i++ {
		createPostVK("https://thiscatdoesnotexist.com/", getText(), t)
		t = t.Add(time.Hour * 4)
		fmt.Println(t)
	}

}
*/
func getText(text string) string {
	var p string
	var l string

	if text == "CITATA" {
		fileTextUrl := "https://socratify.net/quotes/random"
		doc, err := htmlquery.LoadURL(fileTextUrl)
		if err != nil {
			fmt.Println(err)
		}
		p = htmlquery.FindOne(doc, "//h1[@class='b-quote__text']").FirstChild.Data
		l = "30"
	} else {
		p = text
		l = "10"
	}

	reqBody, err := json.Marshal(map[string]string{
		"lenght": l,
		"prompt": p,
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

	var final string

	if text == "CITATA" {
		final = res["replies"][0][1:]
	} else {
		final = p + " " + res["replies"][0][1:]
	}
	final = strings.ReplaceAll(final, "«", "")
	final = strings.ReplaceAll(final, "‎»", "")
	return final
}

func getLastPostTimeVK() time.Time {
	client, err := vkapi.NewVKClientWithToken(vkToken, &vkapi.TokenOptions{}, true)
	if err != nil {
		fmt.Println(err)
		updateToken()
		//return getLastPostTimeVK()
	}

	params := url.Values{}
	params.Set("filter", "postponed")
	wall, err := client.WallGet(vkGroupID, 100, params)
	if err != nil {
		fmt.Println(err)
	}
	return time.Unix(wall.Posts[len(wall.Posts)-1].Date, 0)
}

func createPostVK(urlImage string, text string, t time.Time) {
	client, err := vkapi.NewVKClientWithToken(vkToken, &vkapi.TokenOptions{}, true)
	photo, err := client.UploadGroupWallPhotos(vkGroupID, []string{urlImage})
	if err != nil {
		fmt.Println(err)
		updateToken()
		//createPostVK(urlImage, text, t)
	}

	params := url.Values{}
	params.Set("attachments", client.GetPhotosString(photo))
	params.Set("publish_date", strconv.FormatInt(t.Unix(), 10))
	client.WallPost(vkGroupID, text, params)
}

func updateToken() {
	fmt.Println("JGGGGG")
}

/*
func createPostTG(b *tb.Bot,urlImage string, text string, t time.Time) {
	b.Send(&tb.Chat{Username: "catextus"}, tb.Photo{
		File:    tb.FromDisk("cat.jpg"),
		Caption: text}, tb.SendOptions{

		})
}
*/
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
