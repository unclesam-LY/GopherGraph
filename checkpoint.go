package GopherGraph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Checkpointer 定义了持久化状态的接口，S 是工作流的强类型状态。
type Checkpointer[S any] interface {
	// Save 保存当前线程快照。
	Save(ctx context.Context, threadID string, thread *Thread[S]) error
	// Load 读取之前保存的线程快照，用于恢复执行。
	Load(ctx context.Context, threadID string) (*Thread[S], error)
}

// FileCheckpointer 实现了基于本地 JSON 文件的状态持久化。
type FileCheckpointer[S any] struct {
	dir string // 存放状态文件的文件夹路径
}

// NewFileCheckpointer 创建一个文件检查点管理器。
func NewFileCheckpointer[S any](dir string) (*FileCheckpointer[S], error) {
	// 确保文件夹存在，不存在则创建
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint dir: %w", err)
	}

	return &FileCheckpointer[S]{dir: dir}, nil
}

// Save 将 Thread 序列化为 JSON 并写入本地文件
func (fc *FileCheckpointer[S]) Save(ctx context.Context, threadID string, thread *Thread[S]) error {
	path := filepath.Join(fc.dir, threadID+".json")

	// 使用 MarshalIndent 方便人工阅读生成的 JSON 文件
	data, err := json.MarshalIndent(thread, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal thread: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Load 从本地文件中读取 JSON 并反序列化为 Thread。
func (fc *FileCheckpointer[S]) Load(ctx context.Context, threadID string) (*Thread[S], error) {
	path := filepath.Join(fc.dir, threadID+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read thread file: %w", err)
	}

	var thread Thread[S]
	if err := json.Unmarshal(data, &thread); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thread: %w", err)
	}

	return &thread, nil
}
