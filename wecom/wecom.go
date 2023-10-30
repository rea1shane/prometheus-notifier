package wecom

import (
	"bytes"
	"encoding/json"
	"github.com/morikuni/failure"
	"io"
	"net/http"
	"strconv"
)

const (
	webhookUrl = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="
	okMsg      = `{"errcode":0,"errmsg":"ok"}`
)

func SendBotMarkdownMsg(botKey, content string) error {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	})
	reqBodyBuffer := bytes.NewBuffer(reqBody)
	resp, err := http.Post(webhookUrl+botKey, "application/json", reqBodyBuffer)
	if err != nil {
		return failure.Wrap(err, failure.Context{
			"Bot key": botKey,
			"Content": content,
		})
	}
	defer resp.Body.Close()

	// 处理请求结果
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(respBody) != okMsg {
		return failure.New(
			failure.StringCode("请求企业微信失败"),
			failure.Context{
				"HTTP code": strconv.Itoa(resp.StatusCode),
				"Body":      string(respBody),
			},
		)
	}
	return nil
}
