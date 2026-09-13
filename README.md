# Quiz Battle · 答题对战

> Go 实战练手项目:网络 + 泛型 + Redis

## 游戏规则

两名玩家加入匹配队列,系统撮合后开战。

- **对局**:默认 5 轮,每轮一道题
- **题目**:服务端从 Redis 题库随机抽题,广播给双方
- **作答**:双方同时作答,服务端判断对错,记录时间
- **计分**:答对得 100 分,按答题时间额外加 0~50 分(抢答奖励)
- **胜负**:5 轮结束后分数高者胜

## 整体架构

| 端口 | 协议 | 用途 |
|---|---|---|
| `:7000` | TCP (raw) | 客户端实时连接(对战消息推送) |
| `:7001` | gRPC | 控制面 API(匹配、查询、题库管理、排行榜) |

### 模块划分

```
quizbattle/
├── cmd/
│   ├── server/
│   │   └── main.go              # 入口,启动 TCP + gRPC 两套服务
│   └── client/
│       └── main.go              # 客户端入口,Walking Skeleton 起步处
├── internal/
│   ├── tcpserver/               # TCP 实时对战服务
│   │   ├── server.go            # 监听、accept、conn 管理
│   │   └── session.go           # 玩家会话、读写循环(协议走 pkg/protocol)
│   ├── grpcapi/                 # gRPC 服务实现
│   │   ├── match_service.go     # 匹配、查询
│   │   ├── question_service.go  # 题目 CRUD
│   │   └── leaderboard.go       # 排行榜
│   ├── game/                    # 核心游戏逻辑
│   │   ├── room.go              # 房间状态机
│   │   ├── round.go             # 单回合逻辑
│   │   ├── scoring.go           # 评分(用泛型)
│   │   └── question.go          # 题目类型(用泛型)
│   ├── storage/                 # Redis 封装
│   │   ├── redis.go             # 连接池
│   │   ├── question.go          # 题库操作
│   │   ├── room.go              # 房间状态
│   │   └── leaderboard.go       # 排行榜
│   └── matcher/                 # 撮合逻辑
│       └── matcher.go           # 队列与配对
├── pkg/
│   └── protocol/                # 共享 TCP 协议(服务端与客户端共用)
│       ├── frame.go             # 二进制帧编解码
│       ├── types.go             # 消息类型常量
│       ├── messages.go          # 消息 JSON 结构
│       └── codec.go             # Encode/Decode 便捷封装
├── proto/
│   └── quizbattle.proto         # gRPC 协议定义
├── questions/                   # 题库 JSON 文件(种子数据)
├── client/                      # [演进中] TUI 客户端设计文档(拆分时机说明)
│   └── README.md
└── scripts/
    └── seed_questions.go        # 向 Redis 灌题
```

## TCP 与 gRPC 的分工

### TCP(端口 7000):实时对战通道

- 客户端连接后,所有"对战中的实时消息"走这里
- 自定义二进制协议:`[消息长度 4B][消息类型 1B][payload]`
- 服务端主动 push(题目广播、得分播报、回合切换)
- 客户端主动发(加入队列、提交答案)
- 长连接 + 心跳

### gRPC(端口 7001):结构化 API

服务列表:
- `MatchService`:`JoinQueue` / `LeaveQueue` / `GetStatus`
- `QuestionService`:`Create` / `List` / `GetRandom`(题库管理)
- `LeaderboardService`:`GetTop` / `GetPlayerRank`
- `AdminService`:`Broadcast`(可选)

### 内部协作

```
TCP 连接 ──┐
           ├──> game.Room ──> storage(Redis) ──> 共享状态
gRPC 调用 ─┘
```

## 泛型应用点

### 题目类型参数化

```go
type Question[T Answer] struct {
    ID      string
    Text    string
    Options []string
    Correct T
}

type Answer interface {
    ~string | ~int
}
```

### 计分函数参数化

```go
type ScoringFunc[T any] func(q Question[T], elapsed time.Duration) int

var SpeedScoring ScoringFunc[string] = func(q Question[string], d time.Duration) int {
    base := 100
    bonus := int(50 * (1 - float64(d)/float64(time.Second*10)))
    return base + bonus
}
```

### 答案校验器

```go
type Validator[T comparable] func(given, correct T) bool
```

## Redis 数据结构

| 数据 | 结构 | Key | 用途 |
|---|---|---|---|
| 题库 | Hash | `questions:bank` | 题目存储 |
| 题目分类 | Set | `questions:cat:math` | 按类别抽题 |
| 匹配队列 | List | `match:queue` | 撮合 |
| 房间状态 | Hash | `room:{roomID}` | 单局游戏状态(TTL 30min) |
| 玩家会话 | Hash | `session:{playerID}`(TTL 5min) | 在线状态 |
| 总排行榜 | Sorted Set | `leaderboard:global` | ZADD / ZRANGE |

### 关键操作示例

```go
// 加入匹配队列
LPUSH match:queue playerID

// 撮合
RPOP match:queue

// 保存房间状态
HSET room:abc123 player1 "..." player2 "..." score1 0 score2 0 round 1
EXPIRE room:abc123 1800

// 更新排行榜
ZINCRBY leaderboard:global 100 playerID
```

## 实现阶段

### 阶段 1:基础设施(1-2 天)

- [ ] 创建 `go.mod`,安装依赖(`go-redis/v9`、`grpc`、`protobuf`)
- [ ] `internal/storage/redis.go`,连接池封装
- [ ] `scripts/seed_questions.go`,灌入测试题
- [ ] 验证:能查询题库

### 阶段 2:TCP 通信(2-3 天)

- [ ] `protocol.go`(消息编解码)
- [ ] `server.go` 监听 + accept
- [ ] `session.go` 读写循环
- [ ] CLI 客户端能连上、能收发消息
- [ ] 心跳机制

**学习点**:`net` 包、goroutine 池、连接管理、自定义协议

### 阶段 3:gRPC 服务(2 天)

- [ ] 写 `quizbattle.proto`
- [ ] 生成 Go 代码
- [ ] `MatchService`(加入队列、查询状态)
- [ ] 用 grpcurl 测试

**学习点**:protobuf、gRPC、服务定义

### 阶段 4:游戏逻辑(2-3 天)

- [ ] `Question[T]` 泛型题
- [ ] `ScoringFunc[T]` 泛型计分
- [ ] `Room` 状态机(等待 → 出题 → 作答 → 下一题 → 结束)
- [ ] TCP 与 game 模块打通

**学习点**:类型参数、状态机、并发同步

### 阶段 5:撮合与完整流程(1-2 天)

- [ ] `Matcher`
- [ ] TCP 进入队列 → 撮合 → 创建房间 → 开始对战 → 结束
- [ ] gRPC 排行榜查询

**学习点**:Redis 队列、原子操作

### 阶段 6:打磨(1-2 天)

- [ ] 错误处理、超时、断线重连
- [ ] 单元测试(scoring、validator 这些泛型函数)
- [ ] README + 启动文档

## 技术栈

| 类别 | 选型 |
|---|---|
| 语言 | Go 1.22+ |
| Redis 客户端 | `github.com/redis/go-redis/v9` |
| gRPC | `google.golang.org/grpc` + `protobuf` |
| 协议 | 自定义二进制(TCP) + protobuf(gRPC) |
| 客户端 | 终端 CLI |
| 部署 | `docker-compose up redis` + `go run ./cmd/server` |

## 学习成果

完成后你会熟悉:

### Go 网络
- `net.Listen` / `net.Conn` / `bufio.Reader/Writer`
- 自定义协议(长度前缀、分帧、心跳)
- 高并发连接管理
- TCP 与 gRPC 两种风格对比

### Go 泛型
- 类型参数在业务代码中的应用
- 类型约束(`comparable`、`~int` 等)
- 泛型函数 vs 泛型类型

### Redis
- 5+ 种数据结构(Hash / List / Set / Sorted Set / String)的实际应用
- TTL、原子操作、连接池
- Redis key 命名空间设计

### 工程能力
- 多协议服务共存架构
- 状态机设计
- 项目目录组织

## 启动顺序

1. 启动 Redis:`docker run -d -p 6379:6379 redis:7-alpine`
2. 灌题:`go run ./scripts/seed_questions.go`
3. 启动服务:`go run ./cmd/server`
4. 启动两个客户端分别连 `:7000`,输入用户名,自动入队

## 待办

- [ ] 选择题库领域(通用 / 数学 / IT / 历史)
- [x] 客户端 UI 风格:bubbletea TUI + bubbles/textinput
- [x] 协议共享:`pkg/protocol`(服务端 `internal/tcpserver` 与 `client/` 共用)
- [x] 断线重连:指数退避(1s → 30s),重连后自动重发 JoinQueue
- [ ] 是否需要 Docker(本地装 Redis vs docker-compose)
