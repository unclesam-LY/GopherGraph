# GopherGraph

A lightweight, high-performance, and type-safe Multi-Agent workflow orchestration engine written in Go.  
基于 Go 语言（泛型）实现的轻量级、高性能、类型安全的多智能体（Multi-Agent）协同与任务编排引擎。  
Go 言語（ジェネリクス）で書かれた、軽量・高性能・型安全なマルチエージェント・ワークフロー協調エンジン。

---

* [简体中文](#-简体中文)
* [English](#-english)
* [日本語](#-日本語)

---

## 🇨🇳 简体中文

### 简介
`GopherGraph` 是一个用 Go 语言编写的代码即图（Code-as-Graph）智能体编排引擎，设计灵感来源于 Python 生态的 `LangGraph`。它利用 Go 1.18+ 的泛型机制，让工作流的上下文状态在编译期就具备强类型约束，并依靠 Go 原生的 Goroutines 和 Channels 实现极高并发的本地数据流转。它特别适合构建包含复杂循环（Looping）、并发节点执行（Parallel Execution）和人机协同（Human-in-the-Loop）的智能体工作流。

### 核心特性
- **强类型状态管理：** 利用 Go 泛型定义全局状态 `Graph[S]`，彻底告别松散的 `map[string]any` 和繁琐的类型断言。
- **支持循环结构：** 突破了传统 DAG（有向无环图）工作流的限制，原生支持节点间的双向循环和条件路由。
- **并发分支与合并：** 原生支持多个 Agent 节点的并发执行（基于 `sync.WaitGroup` 调度），并提供线程安全的“分流值拷贝”与“汇合状态合并（Merger）”机制。
- **人机协同中断 (HITL)：** 支持在特定节点前自动挂起，返回执行线程快照，等待人工修改状态或审批通过后一键恢复 (`Resume`)。
- **开箱即用持久化：** 提供 `Checkpointer[S]` 接口与内置的 `FileCheckpointer[S]`，支持将进度保存为 JSON 文件，实现停机或重启后进度无损恢复。
- **纯粹的高性能：** 基于内存流转，零外部依赖，极低的上下文切换开销。

### 快速开始

#### 1. 定义您的状态 (State)
```go
type MyState struct {
    Query    string
    Response string
    Value    int
    Log      []string
}
```

#### 2. 并发分支与合并示例
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    g := GopherGraph.NewGraph[MyState]()

    // 注册节点
    g.AddNode("start", startNode)
    g.AddNode("task1", task1Node)
    g.AddNode("task2", task2Node)
    g.AddNode("end", endNode)

	// 定义合并函数：将并发分支状态合并到主状态中
	merger := func(ctx context.Context, parent MyState, branches []MyState) (MyState, error) {
		for _, b := range branches {
			parent.Value += b.Value
		}
		return parent, nil
	}

    // 建立并发边：从 start 并发分流到 task1 和 task2，执行完后通过 merger 合并状态并去往 end 节点
    g.AddParallelEdges("start", []string{"task1", "task2"}, "end", merger)

    cg, _ := g.Compile()
    thread, _ := cg.Start(context.Background(), "start", MyState{})
}
```

#### 3. 文件持久化 (Checkpointer) 示例
```go
// 创建一个文件存储器，指定存放状态文件的目录
fc, _ := GopherGraph.NewFileCheckpointer[MyState]("./checkpoints")

// 在遇到中断挂起时，将 thread 保存到磁盘
sessionID := "user-session-123"
fc.Save(context.Background(), sessionID, thread)

// 重启程序后，可以从磁盘重新加载进度并恢复执行
loadedThread, _ := fc.Load(context.Background(), sessionID)
thread, _ = cg.Resume(context.Background(), loadedThread, modifiedState)
```

#### 运行内置示例
项目包含了一个完整的“AI翻译 -> 质量检测 -> 人工审核 -> 发布”的交互式 Demo：
```bash
go run examples/translation/main.go
```

#### 运行单元测试
```bash
go test -v ./...
```

---

## 🇺🇸 English

### Introduction
`GopherGraph` is a Code-as-Graph agent orchestration engine built in Go, inspired by Python's `LangGraph`. By leveraging Go 1.18+ Generics, GopherGraph ensures that workflow states are strictly typed at compile-time. Powered by native Goroutines and Channels, it executes agent communication with microsecond latency, making it the perfect engine for building complex looping agent workflows with Human-in-the-Loop (HITL), concurrent branches, and state persistence requirements.

### Key Features
- **Strictly Typed State:** Bind your custom struct to `Graph[S]` via Go generics. Say goodbye to unsafe `map[string]any` and runtime type assertions.
- **Support for Cycles/Loops:** Unlike traditional DAG (Directed Acyclic Graph) engines, GopherGraph natively supports loops and dynamic routing based on agent evaluation.
- **Parallel Branching & Merging:** Run multiple agent nodes concurrently using `sync.WaitGroup`. Thread-safety is achieved through value-copying on fan-out and a custom `ParallelMergeFn` on fan-in.
- **Human-in-the-Loop (HITL):** Pause workflow execution *before* a designated node, capture a snapshot (`Thread`), modify the state, and `Resume` seamlessly.
- **State Persistence (Checkpointer):** Generic `Checkpointer[S]` interface with built-in `FileCheckpointer[S]` for saving and loading execution snapshots to/from JSON files.
- **Zero-Dependency & High-Performance:** Written in pure Go with zero external dependencies, leveraging in-memory queues for lightning-fast orchestration.

### Quick Start

#### 1. Define Your State
```go
type MyState struct {
    Query    string
    Response string
    Value    int
    Log      []string
}
```

#### 2. Parallel execution example
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    g := GopherGraph.NewGraph[MyState]()

    g.AddNode("start", startNode)
    g.AddNode("task1", task1Node)
    g.AddNode("task2", task2Node)
    g.AddNode("end", endNode)

    // Define how to merge states from concurrent branches
    merger := func(ctx context.Context, parent MyState, branches []MyState) (MyState, error) {
        for _, b := range branches {
            parent.Value += b.Value
        }
        return parent, nil
    }

    // Branch from start to task1 and task2 in parallel, merge and transition to end
    g.AddParallelEdges("start", []string{"task1", "task2"}, "end", merger)

    cg, _ := g.Compile()
    thread, _ := cg.Start(context.Background(), "start", MyState{})
}
```

#### 3. State Persistence (Checkpointer) Example
```go
// Initialize a local directory storage
fc, _ := GopherGraph.NewFileCheckpointer[MyState]("./checkpoints")

// Save the thread snapshot when paused
sessionID := "user-session-123"
fc.Save(context.Background(), sessionID, thread)

// Reload the snapshot (e.g. after a process restart) and resume
loadedThread, _ := fc.Load(context.Background(), sessionID)
thread, _ = cg.Resume(context.Background(), loadedThread, modifiedState)
```

#### Run the Interactive Demo
An interactive "Translation -> Review -> Human Approval -> Publish" demo is included:
```bash
go run examples/translation/main.go
```

#### Run Unit Tests
```bash
go test -v ./...
```

---

## 🇯🇵 日本語

### 概要
`GopherGraph` は、Python エコシステムの `LangGraph` に着想を得て開発された、Go 言語向けの Code-as-Graph 型エージェントオーケストレーションエンジンです。Go 1.18+ のジェネリクスを活用することで、ワークフローの状態（State）をコンパイル時に厳密に型定義できます。Go 純正の Goroutines と Channels を利用したメモリ内メッセージングにより、高い並行性と極めて低いレイテンシを実現しています。自律的なループ（Looping）や、並行処理（Parallel Execution）、人間参加型（Human-in-the-Loop）の意思決定を伴うエージェントワークフローの開発に最適です。

### 主な特徴
- **型安全な状態管理:** ジェネリクスを用いて `Graph[S]` に構造体をバインドします。冗長な `map[string]any` やランタイム時の型アサーションから解放されます。
- **ループと条件付き分岐のサポート:** 従来の DAG（有向非巡回グラフ）の制限を超え、エージェントの判定に基づく条件付きルートや双方向のループをサポート。
- **並行処理とマージ（分流と合流）:** `sync.WaitGroup` に基づく複数エージェントノードの並行実行に対応。分流時の値渡しコピーによるスレッドセーフの確保と、合流時の `ParallelMergeFn` による状態マージを実現。
- **Human-in-the-Loop (HITL) の一時停止と再開:** 特定のノードの実行前に処理を自動停止し、スナップショット（`Thread`）を返します。承認や状態修正の後にシームレスに処理を再開（`Resume`）できます。
- **状態の永続化 (Checkpointer):** 抽象的な `Checkpointer[S]` インタフェースと、JSONファイルへの書き出し・読み込みを行う組み込みの `FileCheckpointer[S]` をサポート。サーバー再起動後の進捗復旧が可能。
- **ピュア Go & 高性能:** 外部依存関係ゼロ。メモリ内チャネルを用いた高速なコンテキスト切り替え。

### クイックスタート

#### 1. 状態（State）の定義
```go
type MyState struct {
    Query    string
    Response string
    Value    int
    Log      []string
}
```

#### 2. 並行処理（Parallel）のコード例
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    g := GopherGraph.NewGraph[MyState]()

    g.AddNode("start", startNode)
    g.AddNode("task1", task1Node)
    g.AddNode("task2", task2Node)
    g.AddNode("end", endNode)

    // 並行分岐したノードの状態をマージする関数を定義
    merger := func(ctx context.Context, parent MyState, branches []MyState) (MyState, error) {
        for _, b := range branches {
            parent.Value += b.Value
        }
        return parent, nil
    }

    // startからtask1とtask2を並行実行し、mergerでマージしてendノードに遷移
    g.AddParallelEdges("start", []string{"task1", "task2"}, "end", merger)

    cg, _ := g.Compile()
    thread, _ := cg.Start(context.Background(), "start", MyState{})
}
```

#### 3. 状態の永続化（Checkpointer）のコード例
```go
// 保存先ディレクトリを指定してローカルファイルチェッカーを初期化
fc, _ := GopherGraph.NewFileCheckpointer[MyState]("./checkpoints")

// 一時停止時にスレッドスナップショットをファイルに保存
sessionID := "user-session-123"
fc.Save(context.Background(), sessionID, thread)

// スナップショットを読み込み、プロセス再起動後に処理を再開
loadedThread, _ := fc.Load(context.Background(), sessionID)
thread, _ = cg.Resume(context.Background(), loadedThread, modifiedState)
```

#### インタラクティブデモの実行
「AI翻訳 -> 監査 -> 人間による確認 -> 公開」の一連のループを体験できるデモを用意しています：
```bash
go run examples/translation/main.go
```

#### ユニットテストの実行
```bash
go test -v ./...
```
