package send

import (
	"context"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sirupsen/logrus"
	"os"
)

func Notify(appID, appSecret, fileName, filePath, message string) error {
	// 创建 Client
	client := lark.NewClient(appID, appSecret)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("Error opening file %s: %v", filePath, err)
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Warnf("Error closing file %s: %v", filePath, err)
		}
	}()

	// 创建文件上传请求对象
	createFileReq := larkim.NewCreateFileReqBuilder().
		Body(larkim.NewCreateFileReqBodyBuilder().
			FileType(`pdf`).
			FileName(fileName).
			File(file).
			Build()).
		Build()

	// 发起文件上传请求
	createFileResp, err := client.Im.File.Create(context.Background(), createFileReq)
	if err != nil {
		logrus.Errorf("Error creating file in Lark: %v", err)
		return fmt.Errorf("failed to create file in Lark: %v", err)
	}

	// 服务端错误处理
	if !createFileResp.Success() {
		logrus.Errorf("Server error creating file: Code=%d, Msg=%s, RequestID=%s", createFileResp.Code, createFileResp.Msg, createFileResp.RequestId())
		return fmt.Errorf("server error creating file: Code=%d, Msg=%s, RequestID=%s", createFileResp.Code, createFileResp.Msg, createFileResp.RequestId())
	}

	logrus.Infof("File created successfully: %s", larkcore.Prettify(createFileResp))

	// 创建列表聊天请求对象
	listChatReq := larkim.NewListChatReqBuilder().
		SortType(`ByCreateTimeAsc`).
		PageSize(20).
		Build()

	// 发起列表聊天请求
	listChatResp, err := client.Im.Chat.List(context.Background(), listChatReq)
	if err != nil {
		logrus.Errorf("Error listing chats in Lark: %v", err)
		return fmt.Errorf("failed to list chats in Lark: %v", err)
	}

	// 服务端错误处理
	if !listChatResp.Success() {
		logrus.Errorf("Server error listing chats: Code=%d, Msg=%s, RequestID=%s", listChatResp.Code, listChatResp.Msg, listChatResp.RequestId())
		return fmt.Errorf("server error listing chats: Code=%d, Msg=%s, RequestID=%s", listChatResp.Code, listChatResp.Msg, listChatResp.RequestId())
	}

	logrus.Infof("Chats listed successfully: %s", larkcore.Prettify(listChatResp))

	// 循环发送消息到每个聊天
	for _, i := range listChatResp.Data.Items {
		// 发送文本消息
		createMessageReq := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(`chat_id`).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(*i.ChatId).
				MsgType(`text`).
				Content(fmt.Sprintf("{\"text\":\"%s\"}", message)).
				Build()).
			Build()

		createMessageResp, err := client.Im.Message.Create(context.Background(), createMessageReq)
		if err != nil {
			logrus.Errorf("Error sending text message to chat %s: %v", *i.ChatId, err)
			return fmt.Errorf("failed to send text message to chat %s: %v", *i.ChatId, err)
		}

		if !createMessageResp.Success() {
			logrus.Errorf("Server error sending text message: Code=%d, Msg=%s, RequestID=%s", createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
			return fmt.Errorf("server error sending text message: Code=%d, Msg=%s, RequestID=%s", createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
		}

		logrus.Infof("Text message sent successfully to chat %s", *i.ChatId)

		// 发送文件消息
		createFileMessageReq := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(`chat_id`).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(*i.ChatId).
				MsgType(`file`).
				Content("{\"file_key\":\"" + *createFileResp.Data.FileKey + "\"}").
				Build()).
			Build()

		createFileMessageResp, err := client.Im.Message.Create(context.Background(), createFileMessageReq)
		if err != nil {
			logrus.Errorf("Error sending file message to chat %s: %v", *i.ChatId, err)
			return fmt.Errorf("failed to send file message to chat %s: %v", *i.ChatId, err)
		}

		if !createFileMessageResp.Success() {
			logrus.Errorf("Server error sending file message: Code=%d, Msg=%s, RequestID=%s", createFileMessageResp.Code, createFileMessageResp.Msg, createFileMessageResp.RequestId())
			return fmt.Errorf("server error sending file message: Code=%d, Msg=%s, RequestID=%s", createFileMessageResp.Code, createFileMessageResp.Msg, createFileMessageResp.RequestId())
		}

		logrus.Infof("File message sent successfully to chat %s", *i.ChatId)
	}

	return nil
}
