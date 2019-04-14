package puller

import (
	"math"

	"github.com/u-robot/go-tdlib/client"
)

// Chats returns channels to listen chats and errors.
func Chats(tdlibClient *client.Client) (chan *client.Chat, chan error) {
	chatChan := make(chan *client.Chat, 10)
	errChan := make(chan error, 1)

	var offsetOrder client.Int64JSON = math.MaxInt64
	var offsetChatID int64
	var limit int32 = 100

	go chats(tdlibClient, chatChan, errChan, offsetOrder, offsetChatID, limit)

	return chatChan, errChan
}

func chats(tdlibClient *client.Client, chatChan chan *client.Chat, errChan chan error, offsetOrder client.Int64JSON, offsetChatID int64, limit int32) {
	defer func() {
		close(chatChan)
		close(errChan)
	}()

	for {
		chats, err := tdlibClient.GetChats(&client.GetChatsRequest{
			OffsetOrder:  offsetOrder,
			OffsetChatID: offsetChatID,
			Limit:        limit,
		})
		if err != nil {
			errChan <- err

			return
		}

		if len(chats.ChatIDs) == 0 {
			errChan <- ErrEndOfPull

			break
		}

		for _, chatID := range chats.ChatIDs {
			chat, err := tdlibClient.GetChat(&client.GetChatRequest{
				ChatID: chatID,
			})
			if err != nil {
				errChan <- err

				return
			}

			offsetOrder = chat.Order
			offsetChatID = chat.ID

			chatChan <- chat
		}
	}
}
