package GopherGraph

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestState 是测试用的强类型状态
type TestState struct {
	Value int
	Log   []string
}

// TestSequentialExecution 测试最基础的线性工作流 (A -> B)
func TestSequentialExecution(t *testing.T) {
	g := NewGraph[TestState]()

	g.AddNode("A", func(ctx context.Context, s TestState) (TestState, error) {
		s.Value += 1
		s.Log = append(s.Log, "A")
		return s, nil
	})
	g.AddNode("B", func(ctx context.Context, s TestState) (TestState, error) {
		s.Value *= 2
		s.Log = append(s.Log, "B")
		return s, nil
	})

	g.AddEdge("A", "B")

	cg, err := g.Compile()
	if err != nil {
		t.Fatalf("编译图失败: %v", err)
	}

	thread, err := cg.Start(context.Background(), "A", TestState{Value: 5})
	if err != nil {
		t.Fatalf("启动图失败: %v", err)
	}

	// 验证最终状态
	if !thread.IsFinished {
		t.Errorf("期望工作流运行结束，实际未结束")
	}
	// 计算公式: (5 + 1) * 2 = 12
	if thread.State.Value != 12 {
		t.Errorf("期望 Value 为 12，实际为 %d", thread.State.Value)
	}
	if len(thread.State.Log) != 2 || thread.State.Log[0] != "A" || thread.State.Log[1] != "B" {
		t.Errorf("执行路径日志不匹配，实际为: %v", thread.State.Log)
	}
}

// TestConditionalRouting 测试路由条件分支跳转 (start -> even/odd)
func TestConditionalRouting(t *testing.T) {
	g := NewGraph[TestState]()

	g.AddNode("start", func(ctx context.Context, s TestState) (TestState, error) {
		s.Log = append(s.Log, "start")
		return s, nil
	})
	g.AddNode("even", func(ctx context.Context, s TestState) (TestState, error) {
		s.Log = append(s.Log, "even")
		return s, nil
	})
	g.AddNode("odd", func(ctx context.Context, s TestState) (TestState, error) {
		s.Log = append(s.Log, "odd")
		return s, nil
	})

	g.AddConditionalEdges("start", func(ctx context.Context, s TestState) (string, error) {
		if s.Value%2 == 0 {
			return "even", nil
		}
		return "odd", nil
	})

	cg, err := g.Compile()
	if err != nil {
		t.Fatalf("编译图失败: %v", err)
	}

	// 测试偶数分支
	thread1, err := cg.Start(context.Background(), "start", TestState{Value: 4})
	if err != nil {
		t.Fatalf("启动图失败: %v", err)
	}
	if thread1.State.Log[1] != "even" {
		t.Errorf("偶数测试路由错误，实际路径: %v", thread1.State.Log)
	}

	// 测试奇数分支
	thread2, err := cg.Start(context.Background(), "start", TestState{Value: 7})
	if err != nil {
		t.Fatalf("启动图失败: %v", err)
	}
	if thread2.State.Log[1] != "odd" {
		t.Errorf("奇数测试路由错误，实际路径: %v", thread2.State.Log)
	}
}

// TestInterruptAndResume 测试中断挂起与恢复 (A -> [Interrupt B] -> C)
func TestInterruptAndResume(t *testing.T) {
	g := NewGraph[TestState]()

	g.AddNode("A", func(ctx context.Context, s TestState) (TestState, error) {
		s.Value += 10
		return s, nil
	})
	g.AddNode("B", func(ctx context.Context, s TestState) (TestState, error) {
		s.Value += 20
		return s, nil
	})
	g.AddNode("C", func(ctx context.Context, s TestState) (TestState, error) {
		s.Value += 30
		return s, nil
	})

	g.AddEdge("A", "B")
	g.AddEdge("B", "C")

	// 标记在执行 B 之前进行中断
	g.AddInterrupt("B")

	cg, err := g.Compile()
	if err != nil {
		t.Fatalf("编译图失败: %v", err)
	}

	// 1. 启动图，应该在执行 B 之前停下来
	thread, err := cg.Start(context.Background(), "A", TestState{Value: 0})
	if err != nil {
		t.Fatalf("启动图失败: %v", err)
	}

	// 验证是否成功停在 B 之前
	if !thread.IsPaused {
		t.Errorf("期望工作流处于 Paused 状态")
	}
	if thread.NextNode != "B" {
		t.Errorf("期望下一个节点是 B，实际是 %q", thread.NextNode)
	}
	if thread.State.Value != 10 { // 只执行了 A: 0 + 10 = 10
		t.Errorf("期望 Value 为 10，实际为 %d", thread.State.Value)
	}

	// 2. 模拟人工介入：修改状态值为 100，并调用 Resume 恢复执行
	thread, err = cg.Resume(context.Background(), thread, TestState{Value: 100})
	if err != nil {
		t.Fatalf("恢复图失败: %v", err)
	}

	// 验证是否顺利走完剩余的 B 和 C
	if !thread.IsFinished {
		t.Errorf("期望工作流已结束")
	}
	// 计算公式: B 执行 (100 + 20 = 120) -> C 执行 (120 + 30 = 150)
	if thread.State.Value != 150 {
		t.Errorf("期望最终 Value 为 150，实际为 %d", thread.State.Value)
	}
}

// TestTimeoutCancellation 测试通过 Context 对长时间运行的节点进行超时退出控制
func TestTimeoutCancellation(t *testing.T) {
	g := NewGraph[TestState]()

	g.AddNode("A", func(ctx context.Context, s TestState) (TestState, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			s.Value = 1
			return s, nil
		case <-ctx.Done():
			return s, ctx.Err()
		}
	})

	cg, err := g.Compile()
	if err != nil {
		t.Fatalf("编译图失败: %v", err)
	}

	// 创建一个 20 毫秒的超短超时 Context，而节点 A 需要 100 毫秒才能跑完
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = cg.Start(ctx, "A", TestState{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("期望返回 context.DeadlineExceeded 错误，实际返回: %v", err)
	}
}
