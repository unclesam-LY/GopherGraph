package GopherGraph

import "context"

// NodeFn 代表图中的一个执行节点（Agent）。
// 它接收当前的上下文和状态 S，处理后返回更新后的状态 S，或者返回错误。
type NodeFn[S any] func(ctx context.Context, state S) (S, error)

// RouterFn 代表条件路由函数。
// 它根据当前状态 S，动态决定下一个要执行的节点名称（例如返回 "translator" 或 "reviewer"
type RouterFn[S any] func(ctx context.Context, state S) (string, error)

// Graph 是图的构建器，S 代表用户自定义的状态结构体。
type Graph[S any] struct {
	nodes       map[string]NodeFn[S]   // 节点名称 -> 节点执行函数
	edges       map[string]string      // 普通单向边：源节点 -> 目标节点
	conditional map[string]RouterFn[S] // 条件路由边：源节点 -> 路由函数
	interrupts  map[string]bool        // 标记需要中断的节点：执行完该节点后挂起
}

// NewGraph 创建并初始化一个全新强类型的图构建器
func NewGraph[S any]() *Graph[S] {
	return &Graph[S]{
		nodes:       make(map[string]NodeFn[S]),
		edges:       make(map[string]string),
		conditional: make(map[string]RouterFn[S]),
		interrupts:  make(map[string]bool),
	}
}

// AddNode 向图中注册一个节点（Agent）
func (g *Graph[S]) AddNode(name string, fn NodeFn[S]) {
	g.nodes[name] = fn
}

// AddEdge 建立一条从 from 节点到 to 节点的静态连接线
func (g *Graph[S]) AddEdge(from, to string) {
	g.edges[from] = to
}

// AddConditionalEdges 建立一条从 from 节点出发的条件路由
// 到底去哪，由 router 函数在运行时根据当前 State 动态决定
func (g *Graph[S]) AddConditionalEdges(from string, router RouterFn[S]) {
	g.conditional[from] = router
}

// AddInterrupt 标记在执行 nodeName 节点“之前”进行中断挂起，等待人工确认
func (g *Graph[S]) AddInterrupt(nodeName string) {
	g.interrupts[nodeName] = true
}
