package eino

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"context"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// scrubDirs 是在 copyOutput 阶段需要从宿主 OutputDir 中删除的目录名（沙箱内的中间产物 / 输入）。
var scrubDirs = map[string]bool{
	"skills": true,
	"input":  true,
	"tmp":    true,
}

// copyOutput 把沙箱内的输出目录复制到宿主 OutputDir，并清理隐藏文件与中间目录。
// 沙箱内的 output 子目录会被「拍平」到宿主 OutputDir 根部。
func (r *Runner) copyOutput(ctx context.Context) error {
	log.Infof("%s copyOutput start", r.logPrefix)

	if err := r.sb.CopyFromSandbox(ctx, r.req.OutputDir); err != nil {
		log.Errorf("%s copyOutput CopyFromSandbox failed: %v", r.logPrefix, err)
		return fmt.Errorf("failed to copy output from workspace: %w", err)
	}

	entries, err := os.ReadDir(r.req.OutputDir)
	if err != nil {
		log.Errorf("%s copyOutput ReadDir failed: %v", r.logPrefix, err)
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(r.req.OutputDir, entry.Name())

		switch {
		case strings.HasPrefix(entry.Name(), "."):
			log.Infof("%s copyOutput removing hidden file: %s", r.logPrefix, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("%s copyOutput remove hidden file failed: %s, err: %v", r.logPrefix, entry.Name(), err)
				return fmt.Errorf("failed to remove hidden file %s: %w", entry.Name(), err)
			}

		case entry.IsDir() && scrubDirs[entry.Name()]:
			log.Infof("%s copyOutput removing %s dir", r.logPrefix, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("%s copyOutput remove %s dir failed: %v", r.logPrefix, entry.Name(), err)
				return fmt.Errorf("failed to remove %s directory: %w", entry.Name(), err)
			}

		case entry.IsDir() && entry.Name() == "output":
			log.Infof("%s copyOutput flattening output subdir", r.logPrefix)
			if err := flattenDir(entryPath, r.req.OutputDir); err != nil {
				log.Errorf("%s copyOutput flatten failed: %v", r.logPrefix, err)
				return fmt.Errorf("failed to flatten output directory: %w", err)
			}
		}
	}

	log.Infof("%s copyOutput completed", r.logPrefix)
	return nil
}

// flattenDir 把 src 目录中的所有顶层条目移动到 dst，然后删除空的 src 目录。
func flattenDir(src, dst string) error {
	log.Infof("[flattenDir] start src=%s dst=%s", src, dst)

	subEntries, err := os.ReadDir(src)
	if err != nil {
		log.Errorf("[flattenDir] ReadDir failed: %v", err)
		return fmt.Errorf("failed to read dir %s: %w", src, err)
	}

	for _, sub := range subEntries {
		srcPath := filepath.Join(src, sub.Name())
		dstPath := filepath.Join(dst, sub.Name())
		if err := os.Rename(srcPath, dstPath); err != nil {
			log.Errorf("[flattenDir] move %s failed: %v", sub.Name(), err)
			return fmt.Errorf("failed to move %s: %w", sub.Name(), err)
		}
	}

	if err := os.Remove(src); err != nil {
		log.Errorf("[flattenDir] remove src dir failed: %v", err)
		return err
	}

	log.Infof("[flattenDir] completed")
	return nil
}
