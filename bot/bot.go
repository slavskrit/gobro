package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var logger = log.New(os.Stdout, "telegram-bot: ", log.LstdFlags)

func main() {
	// token := os.Getenv("TELEGRAM_TOKEN")
	token := "6913024292:AAFbIt-5sZmqIsTdhUK_I6vZd7J9um5Mcvk"
	if token == "" {
		logger.Fatalf("No TELEGRAM_TOKEN found in env")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(MainHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func CreateTempDirectoryForChat(chatID int64) string {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("%v/voice/", chatID))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logger.Printf("Failed to create temp directory, maybe it's exist: %v", err)
	}
	logger.Printf("Directory for chat %v is created: %v", chatID, dir)
	return dir
}

func DownloadFile(url string, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	logger.Printf("File successfully downloaded from: %v \t to: %v", url, filepath)
	return nil
}

func MainHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && update.Message.Voice != nil {
		logger.Printf("Received audio file with duration: %d seconds", update.Message.Voice.Duration)

		filePath := fmt.Sprintf("%v/%v",
			CreateTempDirectoryForChat(update.Message.Chat.ID), update.Message.ID)

		file, err := b.GetFile(ctx, &bot.GetFileParams{
			FileID: update.Message.Voice.FileID,
		})

		if err != nil {
			logger.Printf("Could not get file for the id: %v", update.Message.Voice.FileID)
		}

		link := b.FileDownloadLink(file)
		DownloadFile(link, filePath)
		logger.Printf("File url is: %v", link)
		// pathToAudioFile, err := b.FileDownloadLink(msg.Voice.FileID, tempDir)
		// if err != nil {
		// 	logger.Printf("Error downloading file: %v", err)
		// 	return
		// }
		// logger.Printf("Audio is downloaded to path: %s", pathToAudioFile)
		// handleVoiceMessage(ctx, bot, update.Message)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: update.Message.ID,
		Reaction: []models.ReactionType{
			{Type: models.ReactionTypeTypeEmoji,
				ReactionTypeEmoji: &models.ReactionTypeEmoji{
					Emoji: "🤔",
				}},
		},
	})
}
