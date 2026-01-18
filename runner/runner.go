package runner

import (
	"context"
	"log"
	"os/exec"
	"time"
)

type Runner struct {
	ch  <-chan string
	ctx context.Context
}

func NewRunner(ctx context.Context, ch <-chan string) *Runner {
	return &Runner{
		ch:  ch,
		ctx: ctx,
	}
}

func (r *Runner) Start() {
	log.Println("[RUNNER]> Runtime started")
	for {
		select {
		case <-r.ctx.Done():
			log.Println("[RUNNER]> Shutting down script runner...")
			return
		case script, ok := <-r.ch:
			if !ok {
				log.Println("[RUNNER]> Script channel closed, runner exiting")
				return
			}
			r.scriptRuntime(script)
		}
	}

}

func (r *Runner) scriptRuntime(scriptPath string) {
	// TODO: make the timeout variable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[RUNNER]> script failed: %v\noutput: %s", err, output)
		return
	}
	log.Printf("[RUNNER]> '%s' executed successfully:\n%s", scriptPath, output)
}
