package GopherGraph

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Thread 代表图执行的具体实例快照（线程上下文）
// 它可以被序列化并存储，用来实现中断后的状态恢复
type Thread[S any] struct {
	State      S      // 当前共享的强类型状态数据
	NextNode   string // 下一步即将执行的节点名称
	IsPaused   bool   // 是否因为触发中断而处于暂停状态
	IsFinished bool   // 工作流是否全部结束（即到达了终点）
}

// CompiledGraph 是编译后的只读图，准备好投入运行。
type CompiledGraph[S any] struct {
	nodes       map[string]NodeFn[S]
	edges       map[string]string
	conditional map[string]RouterFn[S]
	interrupts  map[string]bool
	parallels   map[string]parallelStep[S]
}

// Compile 校验图结构的合法性，并生成可运行的 CompiledGraph。
func (g *Graph[S]) Compile() (*CompiledGraph[S], error) {
	if len(g.nodes) == 0 {
		return nil, errors.New("cannot compile graph: graph contains no nodes")
	}

	// 校验静态边 (From -> To) 的起始和目标节点是否存在
	for from, to := range g.edges {
		if _, exists := g.nodes[from]; !exists {
			return nil, fmt.Errorf("compile error: edge origin %q does not exist", from)
		}
		if _, exists := g.nodes[to]; !exists {
			return nil, fmt.Errorf("compile error: edge origin %q does not exist", to)
		}
	}

	// 校验条件路由边的源节点是否存在
	for from := range g.conditional {
		if _, exists := g.nodes[from]; !exists {
			return nil, fmt.Errorf("compile error: conditional edge origin %q does not exist", from)
		}
	}

	// 校验被标记为中断的节点是否存在
	for node := range g.interrupts {
		if _, exists := g.nodes[node]; !exists {
			return nil, fmt.Errorf("compile error: interrupt node %q does not exist", node)
		}
	}

	// 校验并发连线边的有效性
	for from, step := range g.parallels {
		if _, exists := g.nodes[from]; !exists {
			return nil, fmt.Errorf("compile error: parallel edge origin %q does not exist", from)
		}
		if len(step.targets) == 0 {
			return nil, fmt.Errorf("compile error: parallel edge from %q has no targets", from)
		}

		for _, target := range step.targets {
			if _, exists := g.nodes[target]; !exists {
				return nil, fmt.Errorf("compile error: parallel target %q from %q does not exist", target, from)
			}
		}
		if step.next != "" {
			if _, exists := g.nodes[step.next]; !exists {
				return nil, fmt.Errorf("compile error: parallel next node %q from %q does not exist", step.next, from)
			}
		}
		if step.merger == nil {
			return nil, fmt.Errorf("compile error: parallel step from %q has nil merger", from)
		}
	}

	// 返回编译好的只读图
	return &CompiledGraph[S]{
		nodes:       g.nodes,
		edges:       g.edges,
		conditional: g.conditional,
		interrupts:  g.interrupts,
		parallels:   g.parallels,
	}, nil
}

// Start 从指定的起始节点开始运行图，并持续流转，直到遇到中断或运行结束
func (cg *CompiledGraph[S]) Start(ctx context.Context, startNode string, initialState S) (*Thread[S], error) {
	thread := &Thread[S]{
		State:    initialState,
		NextNode: startNode,
	}

	return cg.run(ctx, thread)
}

// Resume 恢复执行一个被暂停的线程，并允许注入外部修改后的状态数据（例如人工审批修改后的结果）
func (cg *CompiledGraph[S]) Resume(ctx context.Context, thread *Thread[S], modifiedState S) (*Thread[S], error) {
	if !thread.IsPaused {
		return nil, errors.New("cannot resume: thread is not paused")
	}
	if thread.IsFinished {
		return nil, errors.New("cannot resume: thread is already finished")
	}

	// 注入人工修改后的状态，并解除暂停标记
	thread.State = modifiedState
	thread.IsPaused = false

	return cg.run(ctx, thread)
}

// run 是引擎内部循环调度器，负责驱动节点向前流转
func (cg *CompiledGraph[S]) run(ctx context.Context, thread *Thread[S]) (*Thread[S], error) {
	for {
		// 检查 Context，支持外部超时控制或取消
		if err := ctx.Err(); err != nil {
			return thread, err
		}

		currentNodeName := thread.NextNode
		// 如果下一个执行节点为空，说明整条流水线已运行结束
		if currentNodeName == "" {
			thread.IsFinished = true
			return thread, nil
		}

		// 查找并执行当前节点
		nodeFn, exists := cg.nodes[currentNodeName]
		if !exists {
			return thread, fmt.Errorf("runtime error: node %q not found", currentNodeName)
		}

		newState, err := nodeFn(ctx, thread.State)
		if err != nil {
			return thread, fmt.Errorf("node %q execution error: %w", currentNodeName, err)
		}
		thread.State = newState

		// 计算下一个该执行的节点名称
		var nextNode string
		if step, isParallel := cg.parallels[currentNodeName]; isParallel {
			// 【处理并发分流】
			var wg sync.WaitGroup
			branches := make([]S, len(step.targets))
			errs := make([]error, len(step.targets))
			for i, target := range step.targets {
				wg.Add(1)
				// 启动并发协程运行分支节点。传入 thread.State 的副本（Go 默认是值拷贝）
				go func(idx int, targetNode string, stateCopy S) {
					defer wg.Done()
					if ctx.Err() != nil {
						errs[idx] = ctx.Err()
						return
					}
					nodeFn := cg.nodes[targetNode]
					resState, err := nodeFn(ctx, stateCopy)
					if err != nil {
						errs[idx] = err
						return
					}
					branches[idx] = resState
				}(i, target, thread.State)
			}
			// 等待所有分支协程执行完毕
			wg.Wait()
			// 检查是否有任何分支报错
			for _, err := range errs {
				if err != nil {
					return thread, fmt.Errorf("parallel branch execution error: %w", err)
				}
			}
			// 【状态合并】调用用户自定义的合并函数
			mergedState, err := step.merger(ctx, thread.State, branches)
			if err != nil {
				return thread, fmt.Errorf("parallel merger execution error: %w", err)
			}
			thread.State = mergedState
			nextNode = step.next
		} else if routerFn, ok := cg.conditional[currentNodeName]; ok {
			// 如果有条件路由函数，则通过路由函数动态计算去向
			next, err := routerFn(ctx, thread.State)
			if err != nil {
				return thread, fmt.Errorf("router for node %q execution error: %w", currentNodeName, err)
			}
			nextNode = next
		} else {
			// 否则使用静态连线边
			nextNode = cg.edges[currentNodeName]
		}

		// 更新快照中的“下一个节点”
		thread.NextNode = nextNode

		// 如果下一站是终点，直接进入下一次循环触发结束逻辑
		if nextNode == "" {
			continue
		}

		// 核心中断机制：如果即将进入的下一个节点被标记为了中断节点，则在此挂起
		if cg.interrupts[nextNode] {
			thread.IsPaused = true
			return thread, nil // 暂停执行，返回当前快照供外部人工介入
		}
	}
}
