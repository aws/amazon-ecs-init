package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	dataDir             = "/var/lib/ecs/data/"
	logBundleNotifyFile = dataDir + "logscollect"
	collectScript       = "/usr/libexec/amazon-ecs-logs-collector.sh"
	tmpCollectDir       = "/usr/libexec/collect"
)

func (e *Engine) logCollectorTriggerWatcher(ctx context.Context) (string, error) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(logBundleNotifyFile); err == nil {
				os.Remove(logBundleNotifyFile) // todo handle/log errors
				execCtx, cancel := context.WithTimeout(ctx, time.Minute*3)
				defer cancel()
				cmd := exec.CommandContext(execCtx, collectScript)
				out, err := cmd.CombinedOutput()
				if err != nil {
					return string(out), err
				}
				os.RemoveAll(tmpCollectDir) // todo handle/log errors
				matches, err := filepath.Glob("/usr/libexec/collect-i-*.*")
				if len(matches) > 0 {
					filename := filepath.Base(matches[0])
					err := os.Rename(matches[0], dataDir+filename)
					if err != nil {
						return "", err
					}
					return dataDir + filename, nil
				}
				return "", fmt.Errorf("Could not find the log bundle in /usr/libexec")
			}
		case <-ctx.Done():
			return "", nil
		}
	}
}
