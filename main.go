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
	"time"

	"github.com/antchfx/htmlquery"
	vkapi "github.com/himidori/golang-vk-api"
	tb "gopkg.in/tucnak/telebot.v2"
)

const AdminID = 303629013

var (
	RBtnCreatePosts = tb.ReplyButton{Text: "Заготовить посты"}
	ReplyMain       = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ResizeReplyKeyboard: true,
			ReplyKeyboard: [][]tb.ReplyButton{
				{
					RBtnCreatePosts,
				},
			},
		},
	}

	IBtnCreate   = tb.InlineButton{Text: "✔️ Готово", Unique: "ok"}
	IBtnEditText = tb.InlineButton{Text: "✏️ Редактировать текст", Unique: "editText"}
	IBtnReText   = tb.InlineButton{Text: "🔄 Обновить текст", Unique: "reText"}
	IBtnReCat    = tb.InlineButton{Text: "🔄 Обновить кота", Unique: "reCat"}
	InlinePost   = &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{{IBtnCreate}, {IBtnEditText}, {IBtnReText, IBtnReCat}},
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

	poller := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	middle := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		if upd.Message.Sender.ID == AdminID {
			return true
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
		b.Send(m.Sender, "Вот тебе менюшка!", ReplyMain)
	})

	b.Handle(&RBtnCreatePosts, func(m *tb.Message) {
		DownloadFile("cat.jpg", "https://thiscatdoesnotexist.com/")
		text := getText()
		b.Send(m.Sender, tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)

		SuperTimer = getLastPostTimeVK()
	})
	b.Handle(&IBtnCreate, func(c *tb.Callback) {
		SuperTimer = SuperTimer.Add(time.Hour * 4)
		createPostVK("cat.jpg", c.Message.Photo.Caption, SuperTimer)
	})
	b.Handle(&IBtnReText, func(c *tb.Callback) {
		text := getText()
		b.Edit(c.Message, tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: text,
		}, InlinePost)
	})
	b.Handle(&IBtnReCat, func(c *tb.Callback) {
		DownloadFile("cat.jpg", "https://thiscatdoesnotexist.com/")
		b.Edit(c.Message, tb.Photo{
			File:    tb.FromDisk("cat.jpg"),
			Caption: c.Message.Photo.Caption,
		}, InlinePost)
	})
	b.Handle(&IBtnEditText, func(c *tb.Callback) {
		b.Send(c.Sender, "Отправь новый текст")
		IsBotStateEditText = true
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if IsBotStateEditText {
			IsBotStateEditText = false
			b.Send(m.Sender, tb.Photo{
				File:    tb.FromDisk("cat.jpg"),
				Caption: m.Text,
			}, InlinePost)
		}
	})

	b.Start()

}

func CreatePosts() {

	t := time.Date(2021, time.January, 21, 20, 0, 0, 0, time.Local)
	for i := 0; i < 72; i++ {
		createPostVK("https://thiscatdoesnotexist.com/", getText(), t)
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

func getLastPostTimeVK() time.Time {
	var vkToken = "d46b8dcdb65a9796844341b9b94ccca5e6eb649e5fa7f311d54f42ec1d47b15297201a608876b6fb5ec73"
	var vkGroupID = -199800931

	client, err := vkapi.NewVKClientWithToken(vkToken, &vkapi.TokenOptions{}, true)
	if err != nil {
		fmt.Println(err)
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
	var vkToken = "d46b8dcdb65a9796844341b9b94ccca5e6eb649e5fa7f311d54f42ec1d47b15297201a608876b6fb5ec73"
	var vkGroupID = -199800931

	client, err := vkapi.NewVKClientWithToken(vkToken, &vkapi.TokenOptions{}, true)
	photo, err := client.UploadGroupWallPhotos(vkGroupID, []string{urlImage})
	if err != nil {
		fmt.Println(err)
	}

	params := url.Values{}
	params.Set("attachments", client.GetPhotosString(photo))
	params.Set("publish_date", strconv.FormatInt(t.Unix(), 10))
	client.WallPost(vkGroupID, text, params)
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
