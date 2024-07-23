package send

import (
	"context"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"inspection-server/pkg/common"
	"os"
)

func Notify(appID, appSecret string) error {
	// 创建 Client
	client := lark.NewClient(appID, appSecret)
	// 创建请求对象
	file, err := os.Open(common.PrintPDFPath)
	if err != nil {
		return err
	}
	defer file.Close()

	createFileReq := larkim.NewCreateFileReqBuilder().
		Body(larkim.NewCreateFileReqBodyBuilder().
			FileType(`pdf`).
			FileName(`inspection.pdf`).
			File(file).
			Build()).
		Build()

	// 发起请求
	createFileResp, err := client.Im.File.Create(context.Background(), createFileReq)

	// 处理错误
	if err != nil {
		return err
	}

	// 服务端错误处理
	if !createFileResp.Success() {
		fmt.Println(createFileResp.Code, createFileResp.Msg, createFileResp.RequestId())
		return err
	}

	// 业务处理
	fmt.Println(larkcore.Prettify(createFileResp))

	// 创建请求对象
	listChatReq := larkim.NewListChatReqBuilder().
		SortType(`ByCreateTimeAsc`).
		PageSize(20).
		Build()

	// 发起请求
	listChatResp, err := client.Im.Chat.List(context.Background(), listChatReq)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 服务端错误处理
	if !listChatResp.Success() {
		fmt.Println(listChatResp.Code, listChatResp.Msg, listChatResp.RequestId())
		return err
	}

	// 业务处理
	fmt.Println(larkcore.Prettify(listChatResp))

	for _, i := range listChatResp.Data.Items {

		createMessageReq := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(`chat_id`).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(*i.ChatId).
				MsgType(`file`).
				Content("{\"file_key\":\"" + *createFileResp.Data.FileKey + "\"}").
				Build()).
			Build()

		// 发起请求
		createMessageResp, err := client.Im.Message.Create(context.Background(), createMessageReq)

		// 处理错误
		if err != nil {
			fmt.Println(err)
			return err
		}

		// 服务端错误处理
		if !createMessageResp.Success() {
			fmt.Println(createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
			return err
		}

		// 业务处理
		fmt.Println(larkcore.Prettify(createMessageResp))
	}

	return nil
}
