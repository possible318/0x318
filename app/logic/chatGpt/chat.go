package chatGpt

import (
	"context"
	"errors"
	"fmt"
	"github.com/open_tool/app/common"
	"github.com/open_tool/app/utils/logger"
	tiktoken "github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ChatGPT struct {
	client         *openai.Client
	ctx            context.Context
	userId         string
	maxText        int
	maxQuestionLen int
	maxAnswerLen   int
	timeOut        time.Duration // 超时时间, 0表示不超时
	doneChan       chan struct{}
	cancel         func()
	ChatContext    *ChatContext
}

func New(userId string) *ChatGPT {
	var ctx context.Context
	var cancel func()

	ctx, cancel = context.WithTimeout(context.Background(), 600*time.Second)
	timeOutChan := make(chan struct{}, 1)
	go func() {
		<-ctx.Done()
		timeOutChan <- struct{}{} // 发送超时信号，或是提示结束，用于聊天机器人场景，配合GetTimeOutChan() 使用
	}()

	appCfg := common.GetConfigData()
	// token
	token := appCfg.Token
	openConf := openai.DefaultConfig(token)
	// base url
	baseUrl := appCfg.BaseURL
	if baseUrl != "" {
		openConf.BaseURL = baseUrl
	}

	// 代理
	if appCfg.HttpProxy != "" {
		proxyUrl, err := url.Parse(appCfg.HttpProxy)
		if err != nil {
			panic(err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		openConf.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	return &ChatGPT{
		client:         openai.NewClientWithConfig(openConf),
		ctx:            ctx,
		userId:         userId,
		maxQuestionLen: appCfg.MaxQuestionLen, // 最大问题长度
		maxAnswerLen:   appCfg.MaxAnswerLen,   // 最大答案长度
		maxText:        appCfg.MaxText,        // 最大文本 = 问题 + 回答, 接口限制
		timeOut:        appCfg.SessionTimeout,
		doneChan:       timeOutChan,
		cancel: func() {
			cancel()
		},
		ChatContext: NewContext(),
	}
}

func (c *ChatGPT) Close() {
	c.cancel()
}

func (c *ChatGPT) GetDoneChan() chan struct{} {
	return c.doneChan
}

func (c *ChatGPT) SetMaxQuestionLen(maxQuestionLen int) int {
	if maxQuestionLen > c.maxText-c.maxAnswerLen {
		maxQuestionLen = c.maxText - c.maxAnswerLen
	}
	c.maxQuestionLen = maxQuestionLen
	return c.maxQuestionLen
}

func (c *ChatGPT) ChatStream(prompt string) {
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}
	stream, err := c.client.CreateChatCompletionStream(c.ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}

func (c *ChatGPT) ChatCompletion(prompt string) string {

	req := openai.CompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		MaxTokens:   c.maxAnswerLen,
		Prompt:      prompt,
		Temperature: 0.6,
		User:        c.userId,
		Stop:        []string{c.ChatContext.aiRole.Name + ":", c.ChatContext.humanRole.Name + ":"},
	}

	resp, err := c.client.CreateCompletion(c.ctx, req)
	if err != nil {
		logger.Error("Completion error: " + err.Error())
		return err.Error()
	}

	resp.Choices[0].Text = formatAnswer(resp.Choices[0].Text)
	c.ChatContext.old = append(c.ChatContext.old, conversation{
		Role:   c.ChatContext.humanRole,
		Prompt: prompt,
	})
	c.ChatContext.old = append(c.ChatContext.old, conversation{
		Role:   c.ChatContext.aiRole,
		Prompt: resp.Choices[0].Text,
	})
	c.ChatContext.seqTimes++
	return resp.Choices[0].Text
}

func (c *ChatGPT) ChatStreamCompletion(prompt string) {
	req := openai.CompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: c.maxAnswerLen,
		Prompt:    prompt,
		Stream:    true,
	}
	stream, err := c.client.CreateCompletionStream(c.ctx, req)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		fmt.Printf("Stream response: %v\n", response)
	}
}

func (c *ChatGPT) TikToken(content string) int {
	encoding := common.Conf.Model

	tkm, err := tiktoken.EncodingForModel(encoding)
	if err != nil {
		logger.Error("getEncoding: " + err.Error())
		return 0
	}
	// encode
	token := tkm.Encode(content, nil, nil)
	return len(token)
}

func (c *ChatGPT) ChatWithContext(question string) (answer string, err error) {
	if c.TikToken(question) > c.maxQuestionLen {
		return "", OverMaxQuestionLength
	}
	if c.ChatContext.seqTimes >= c.ChatContext.maxSeqTimes {
		if c.ChatContext.maintainSeqTimes {
			c.ChatContext.PollConversation()
		} else {
			return "", OverMaxSequenceTimes
		}
	}
	var promptTable []string
	promptTable = append(promptTable, c.ChatContext.background)
	promptTable = append(promptTable, c.ChatContext.preset)
	for _, v := range c.ChatContext.old {
		if v.Role == c.ChatContext.humanRole {
			promptTable = append(promptTable, "\n"+v.Role.Name+": "+v.Prompt)
		} else {
			promptTable = append(promptTable, v.Role.Name+": "+v.Prompt)
		}
	}
	promptTable = append(promptTable, "\n"+c.ChatContext.restartSeq+question)
	prompt := strings.Join(promptTable, "\n")
	prompt += c.ChatContext.startSeq
	// 删除对话，直到prompt的长度满足条件
	for c.TikToken(prompt) > c.maxText {
		if len(c.ChatContext.old) > 1 { // 至少保留一条记录
			c.ChatContext.PollConversation() // 删除最旧的一条对话
			// 重新构建 prompt，计算长度
			promptTable = promptTable[1:] // 删除promptTable中对应的对话
			prompt = strings.Join(promptTable, "\n") + c.ChatContext.startSeq
		} else {
			break // 如果已经只剩一条记录，那么跳出循环
		}
	}
	if c.TikToken(prompt) > c.maxText-c.maxAnswerLen {
		return "", OverMaxTextLength
	}
	model := common.Conf.Model

	userId := c.userId

	var res string
	// 按照模型选择对应的请求
	if model == openai.GPT3Dot5Turbo0301 ||
		model == openai.GPT3Dot5Turbo ||
		model == openai.GPT4 ||
		model == openai.GPT40314 ||
		model == openai.GPT432K ||
		model == openai.GPT432K0314 {

		res, err = c.chatCompletion(model, prompt, userId)
	} else {
		res, err = c.completion(model, prompt)
	}

	content := formatAnswer(res)
	c.ChatContext.old = append(c.ChatContext.old, conversation{
		Role:   c.ChatContext.humanRole,
		Prompt: prompt,
	})
	c.ChatContext.old = append(c.ChatContext.old, conversation{
		Role:   c.ChatContext.aiRole,
		Prompt: content,
	})
	c.ChatContext.seqTimes++
	return content, err
}

func (c *ChatGPT) chatCompletion(model, prompt string, userId string) (answer string, err error) {
	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   c.maxAnswerLen,
		Temperature: 0.6,
		User:        userId,
	}
	resp, err := c.client.CreateChatCompletion(c.ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *ChatGPT) completion(model, prompt string) (answer string, err error) {
	req := openai.CompletionRequest{
		Model:       model,
		MaxTokens:   c.maxAnswerLen,
		Prompt:      prompt,
		Temperature: 0.6,
		User:        c.userId,
		Stop:        []string{c.ChatContext.aiRole.Name + ":", c.ChatContext.humanRole.Name + ":"},
	}
	resp, err := c.client.CreateCompletion(c.ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Text, nil
}
