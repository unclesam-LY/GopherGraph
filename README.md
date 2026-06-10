# GopherGraph

A lightweight, high-performance, and type-safe Multi-Agent workflow orchestration engine written in Go.  
基于 Go 语言（泛型）实现的轻量级、高性能、类型安全的多智能体（Multi-Agent）协同与任务编排引擎。  
Go 言語（ジェネリクス）で書かれた、軽量・高性能・型安全なマルチエージェント・ワークフロー協調エンジン。

---

* [简体中文](#-简体中文)
* [English](#-english)
* [日本語](#-日本語)

---

## 简体中文

### 简介
`GopherGraph` 是一个用 Go 语言编写的代码即图（Code-as-Graph）智能体编排引擎，设计灵感来源于 Python 生态的 `LangGraph`。它利用 Go 1.18+ 的泛型机制，让工作流的上下文状态在编译期就具备强类型约束，并依靠 Go 原生的 Goroutines 和 Channels 实现极高并发的本地数据流转。它特别适合构建包含复杂循环（Looping）和人机协同（Human-in-the-Loop）的智能体工作流。

### 核心特性
- **强类型状态管理：** 利用 Go 泛型定义全局状态 `Graph[S]`，彻底告别松散的 `map[string]any` 和繁琐的类型断言。
- **支持循环结构：** 突破了传统 DAG（有向无环图）工作流的限制，原生支持节点间的双向循环和条件路由。
- **人机协同中断 (HITL)：** 支持在特定节点前自动挂起，返回执行线程快照，等待人工修改状态或审批通过后一键恢复 (`Resume`)。
- **纯粹的高性能：** 基于内存流转，零外部依赖，极低的上下文切换开销。

### 快速开始

#### 1. 定义您的状态 (State)
```go
type MyState struct {
    Query    string
    Response string
}
```

#### 2. 构建工作流图
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    // 初始化一个强类型的图构建器
    g := GopherGraph.NewGraph[MyState]()

    // 注册节点 (Node)
    g.AddNode("agentA", func(ctx context.Context, s MyState) (MyState, error) {
        s.Response = "Hello from Agent A"
        return s, nil
    })

    // 建立连线 (Edge)
    g.AddEdge("agentA", "anotherNode")
    
    // 设置中断拦截点 (Interrupt)
    g.AddInterrupt("anotherNode")

    // 编译图
    cg, _ := g.Compile()

    // 启动运行
    thread, _ := cg.Start(context.Background(), "agentA", MyState{Query: "Test"})
}
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

## English

### Introduction
`GopherGraph` is a Code-as-Graph agent orchestration engine built in Go, inspired by Python's `LangGraph`. By leveraging Go 1.18+ Generics, GopherGraph ensures that workflow states are strictly typed at compile-time. Powered by native Goroutines and Channels, it executes agent communication with microsecond latency, making it the perfect engine for building complex looping agent workflows with Human-in-the-Loop (HITL) requirements.

### Key Features
- **Strictly Typed State:** Bind your custom struct to `Graph[S]` via Go generics. Say goodbye to unsafe `map[string]any` and runtime type assertions.
- **Support for Cycles/Loops:** Unlike traditional DAG (Directed Acyclic Graph) engines, GopherGraph natively supports loops and dynamic routing based on agent evaluation.
- **Human-in-the-Loop (HITL):** Pause workflow execution *before* a designated node, capture a snapshot (`Thread`), modify the state, and `Resume` seamlessly.
- **Zero-Dependency & High-Performance:** Written in pure Go with zero external dependencies, leveraging in-memory queues for lightning-fast orchestration.

### Quick Start

#### 1. Define Your State
```go
type MyState struct {
    Query    string
    Response string
}
```

#### 2. Define and Run the Graph
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    // Initialize the strongly-typed graph builder
    g := GopherGraph.NewGraph[MyState]()

    // Add nodes (agents)
    g.AddNode("agentA", func(ctx context.Context, s MyState) (MyState, error) {
        s.Response = "Hello from Agent A"
        return s, nil
    })

    // Add edges
    g.AddEdge("agentA", "anotherNode")
    
    // Add interrupts
    g.AddInterrupt("anotherNode")

    // Compile
    cg, _ := g.Compile()

    // Start execution
    thread, _ := cg.Start(context.Background(), "agentA", MyState{Query: "Test"})
}
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

## 日本語

### 概要
`GopherGraph` は、Python エコシステムの `LangGraph` に着想を得て開発された、Go 言語向けの Code-as-Graph 型エージェントオーケストレーションエンジンです。Go 1.18+ のジェネリクスを活用することで、ワークフローの状態（State）をコンパイル時に厳密に型定義できます。Go 純正の Goroutines と Channels を利用したメモリ内メッセージングにより、高い並行性と極めて低いレイテンシを実現しています。自律的なループ（Looping）や人間参加型（Human-in-the-Loop）の意思決定を伴うエージェントワークフローの開発に最適です。

### 主な特徴
- **型安全な状態管理:** ジェネリクスを用いて `Graph[S]` に構造体をバインドします。冗長な `map[string]any` やランタイム時の型アサーションから解放されます。
- **ループと条件付き分岐のサポート:** 従来の DAG（有向非巡回グラフ）の制限を超え、エージェントの判定に基づく条件付きルートや双方向のループをサポート。
- **Human-in-the-Loop (HITL) の一時停止と再開:** 特定のノードの実行前に処理を自動停止し、スナップショット（`Thread`）を返します。承認や状態修正の後にシームレスに処理を再開（`Resume`）できます。
- **ピュア Go & 高性能:** 外部依存関係ゼロ。メモリ内チャネルを用いた高速なコンテキスト切り替え。

### クイックスタート

#### 1. 状態（State）の定義
```go
type MyState struct {
    Query    string
    Response string
}
```

#### 2. ワークフロー図の構築と実行
```go
package main

import (
    "context"
    "fmt"
    "GopherGraph"
)

func main() {
    // 型安全なグラフビルダーの初期化
    g := GopherGraph.NewGraph[MyState]()

    // ノード（エージェント）の追加
    g.AddNode("agentA", func(ctx context.Context, s MyState) (MyState, error) {
        s.Response = "Hello from Agent A"
        return s, nil
    })

    // エッジ（遷移線）の接続
    g.AddEdge("agentA", "anotherNode")
    
    // 割り込み（一時停止）ポイントの設定
    g.AddInterrupt("anotherNode")

    // コンパイル
    cg, _ := g.Compile()

    // 実行開始
    thread, _ := cg.Start(context.Background(), "agentA", MyState{Query: "Test"})
}
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
