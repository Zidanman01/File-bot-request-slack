package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	token := "your slack bot token here"
	api := slack.New(token)

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["type"] == "url_verification" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, data["challenge"])
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				message := strings.ToLower(ev.Text)

				if strings.Contains(message, "request file") {
					parts := strings.Split(message, "request file")
					if len(parts) > 1 {
						fileName := strings.TrimSpace(parts[1])
						fmt.Printf("User meminta file: %s\n", fileName)
						go uploadFile(api, ev.Channel, fileName)
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Server Bot Aktif di Port 3000...")
	http.ListenAndServe(":3000", nil)
}

func uploadFile(api *slack.Client, channelID string, filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		api.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf("File `%s` tidak ditemukan di server.", filePath), false))
		return
	}
	defer f.Close()

	fileInfo, _ := f.Stat()

	params := slack.UploadFileV2Parameters{
		Channel:  channelID,
		File:     filePath,
		Filename: filePath,
		FileSize: int(fileInfo.Size()),
	}

	_, err = api.UploadFileV2(params)
	if err != nil {
		fmt.Printf("Gagal upload: %v\n", err)
		api.PostMessage(channelID, slack.MsgOptionText("Terjadi kesalahan saat mengunggah file.", false))
	} else {
		fmt.Printf("Berhasil mengirim %s\n", filePath)
	}
}
