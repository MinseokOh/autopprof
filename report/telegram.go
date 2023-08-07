package report

import (
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegoutil"
	"io"
	"os"
	"time"
)

var (
	_ Reporter              = &TelegramReporter{}
	_ telegoapi.NamedReader = &telegramFile{}
)

// TelegramReporter is the reporter to send the profiling report to the
// specific Telegram chat id.
type TelegramReporter struct {
	chatId int64
	app    string

	client *telego.Bot
}

// TelegramReporterOption is the option for the Telegram reporter.
type TelegramReporterOption struct {
	ChatId int64
	App    string
	Token  string
}

const (
	cpuTelegramCommentFmt = "[%s] [CPU] usage (*%.2f%%*) > threshold (*%.2f%%*)"
	memTelegramCommentFmt = "[%s] [MEM] usage (*%.2f%%*) > threshold (*%.2f%%*)"
)

// telegramFile is wrapper for telegoapi.NamedReader
type telegramFile struct {
	io.Reader
	name string
}

// Name implements telegoapi.NamedReader
func (t telegramFile) Name() string {
	return t.name
}

// NewTelegramReporter returns the new NewTelegramReporter.
func NewTelegramReporter(opt *TelegramReporterOption) *TelegramReporter {
	bot, err := telego.NewBot(opt.Token)
	if err != nil {
		return nil
	}

	return &TelegramReporter{
		chatId: opt.ChatId,
		app:    opt.App,
		client: bot,
	}
}

// ReportCPUProfile sends the CPU profiling data to the Slack.
func (t *TelegramReporter) ReportCPUProfile(
	ctx context.Context, r io.Reader, ci CPUInfo,
) error {
	hostname, _ := os.Hostname() // Don't care about this error.
	var (
		now      = time.Now().Format(reportTimeLayout)
		filename = fmt.Sprintf(HeapProfileFilenameFmt, t.app, hostname, now)
		comment  = fmt.Sprintf(cpuTelegramCommentFmt, t.app, ci.UsagePercentage, ci.ThresholdPercentage)
	)

	if _, err := t.client.SendDocument(&telego.SendDocumentParams{
		ChatID:   telegoutil.ID(t.chatId),
		Document: telegoutil.File(telegramFile{Reader: r, name: filename}),
		Caption:  comment,
	}); err != nil {
		return fmt.Errorf("autopprof: failed to upload a file to telegram bot: %w", err)
	}
	return nil
}

// ReportHeapProfile sends the heap profiling data to the Slack.
func (t *TelegramReporter) ReportHeapProfile(
	ctx context.Context, r io.Reader, mi MemInfo,
) error {
	hostname, _ := os.Hostname() // Don't care about this error.
	var (
		now      = time.Now().Format(reportTimeLayout)
		filename = fmt.Sprintf(HeapProfileFilenameFmt, t.app, hostname, now)
		comment  = fmt.Sprintf(memTelegramCommentFmt, t.app, mi.UsagePercentage, mi.ThresholdPercentage)
	)

	if _, err := t.client.SendDocument(&telego.SendDocumentParams{
		ChatID:   telegoutil.ID(t.chatId),
		Document: telegoutil.File(telegramFile{Reader: r, name: filename}),
		Caption:  comment,
	}); err != nil {
		return fmt.Errorf("autopprof: failed to upload a file to telegram bot: %w", err)
	}
	return nil
}
