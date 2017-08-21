package progress

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"time"
)

type ProgressConsole struct {
	*progressCount
	progress    *uiprogress.Progress
	progressBar *uiprogress.Bar
}

func NewProgressConsole() *ProgressConsole {
	progress := uiprogress.New()
	progress.RefreshInterval = time.Second
	progress.Start()
	return &ProgressConsole{progressCount: &progressCount{}, progress: progress}
}

func (progressConsole *ProgressConsole) createProgressBar() {
	if progressConsole.getTotalInternal() > 0 {
		fmt.Println("[Applying changes]")
	}
	progressConsole.progressBar = progressConsole.progress.AddBar(progressConsole.getTotalInternal())
	progressConsole.progressBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  [%s: %d/%d]", progressConsole.getStageInternal(), b.Current(), b.Total)
	})
	progressConsole.progressBar.AppendCompleted()
	progressConsole.progressBar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  Time: %s", b.TimeElapsedString())
	})
}

func (progressConsole *ProgressConsole) SetTotal(total int) {
	progressConsole.setTotalInternal(total)
	progressConsole.createProgressBar()
}

func (progressConsole *ProgressConsole) Advance(stage string) {
	progressConsole.advanceInternal(stage)
	progressConsole.progressBar.Incr()
}

func (progressConsole *ProgressConsole) Done() {
	progressConsole.doneInternal()
	progressConsole.progress.Stop()
}

func (progressConsole *ProgressConsole) IsDone() bool {
	return progressConsole.isDoneInternal()
}