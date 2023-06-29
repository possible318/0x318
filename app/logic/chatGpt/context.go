package chatGpt

import (
	"bytes"
	"encoding/gob"
)

// 会话session
type Session struct {
	// 会话id
	Id string
	// 会话上下文
	Context *ChatContext
}

var (
	DefaultAiRole    = "AI"
	DefaultHumanRole = "Human"

	DefaultCharacter  = []string{"helpful", "creative", "clever", "friendly", "lovely", "talkative"}
	DefaultBackground = "The following is a conversation with AI assistant. The assistant is %s"
	DefaultPreset     = "\n%s: 你好，让我们开始愉快的谈话！\n%s: 我是 AI assistant ，请问你有什么问题？"
)

type (
	ChatContext struct {
		background  string // 对话背景
		preset      string // 预设对话
		maxSeqTimes int    // 最大对话次数
		aiRole      *role  // AI角色
		humanRole   *role  // 人类角色

		old        []conversation // 旧对话
		restartSeq string         // 重新开始对话的标识
		startSeq   string         // 开始对话的标识

		seqTimes int // 对话次数

		maintainSeqTimes bool // 是否维护对话次数 (自动移除旧对话)
	}

	ChatContextOption func(*ChatContext)

	conversation struct {
		Role   *role
		Prompt string
	}

	role struct {
		Name string
	}
)

func NewContext(options ...ChatContextOption) *ChatContext {
	ctx := &ChatContext{
		aiRole:           &role{Name: DefaultAiRole},
		humanRole:        &role{Name: DefaultHumanRole},
		background:       "",
		maxSeqTimes:      1000,
		preset:           "",
		old:              []conversation{},
		seqTimes:         0,
		restartSeq:       "\n" + DefaultHumanRole + ": ",
		startSeq:         "\n" + DefaultAiRole + ": ",
		maintainSeqTimes: false,
	}

	for _, option := range options {
		option(ctx)
	}
	return ctx
}

// PollConversation 移除最旧的一则对话
func (c *ChatContext) PollConversation() {
	c.old = c.old[1:]
	c.seqTimes--
}

// ResetConversation 重置对话
func (c *ChatContext) ResetConversation(userid string) {
	c.old = []conversation{}
	c.seqTimes = 0
	// 删除会话记录
}

// SaveConversation 保存对话
func (c *ChatContext) SaveConversation(userid string) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(c.old)
	if err != nil {
		return err
	}

	// 保存到数据库中
	return nil
}

// LoadConversation 加载对话
func (c *ChatContext) LoadConversation(userid string) error {
	// 从数据库中加载

	//dec := gob.NewDecoder(strings.NewReader(public.UserService.GetUserSessionContext(userid)))
	//err := dec.Decode(&c.old)
	//if err != nil {
	//	return err
	//}
	//c.seqTimes = len(c.old)
	return nil
}

func (c *ChatContext) SetHumanRole(role string) {
	c.humanRole.Name = role
	c.restartSeq = "\n" + c.humanRole.Name + ": "
}

func (c *ChatContext) SetAiRole(role string) {
	c.aiRole.Name = role
	c.startSeq = "\n" + c.aiRole.Name + ": "
}

func (c *ChatContext) SetMaxSeqTimes(times int) {
	c.maxSeqTimes = times
}

func (c *ChatContext) GetMaxSeqTimes() int {
	return c.maxSeqTimes
}

func (c *ChatContext) SetBackground(background string) {
	c.background = background
}

func (c *ChatContext) SetPreset(preset string) {
	c.preset = preset
}

func formatAnswer(answer string) string {
	for len(answer) > 0 {
		if answer[:1] == "\n" || answer[0] == ' ' {
			answer = answer[1:]
		} else {
			break
		}
	}
	return answer
}

func WithMaxSeqTimes(times int) ChatContextOption {
	return func(c *ChatContext) {
		c.SetMaxSeqTimes(times)
	}
}

// WithOldConversation 从文件中加载对话
func WithOldConversation(userid string) ChatContextOption {
	return func(c *ChatContext) {
		_ = c.LoadConversation(userid)
	}
}

func WithMaintainSeqTimes(maintain bool) ChatContextOption {
	return func(c *ChatContext) {
		c.maintainSeqTimes = maintain
	}
}
