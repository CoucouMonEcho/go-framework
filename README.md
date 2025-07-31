# Code Practise

---

## 1. Go Web框架

### 1.1. Router路由

- [x] 抽象`Server`接口，支持`HTTPServer`实现；
- [x] 分割`/`构造路由树，支持静态匹配[^1]；
- [x] 支持不同节点类型，实现高级路由：`/*`通配符匹配、`/:id`路径参数。

  [^1]: Gin框架使用了前缀树，查找速度快；但代码过于复杂、因此不考虑。

### 1.2. Context上下文

- [x] 封装`*http.Request`与`http.ResponseWriter`，用于Middleware；
- [x] 存储url参数、路径参数、请求体、表单参数；
- [x] 封装`StringValue`，支持返回值类型转化语法糖。

  <details><summary>示例代码</summary>

  ```
    req, err := ctx.QueryValue("file").String()
  ```

  </details>

### 1.3. Middleware中间件

- [x] 重写路由查找、支持节点级别的Middleware[^2]；
- [x] 接入`AccessLog`记录请求日志；
- [x] 接入`OpenTelemetry`可观测性链路；
- [x] 接入`Prometheus`实现性能监控；
- [x] 接入`Errhandle`返回错误页面；
- [x] 接入`Recover`支持从错误中恢复。

  [^2]: 匹配路由二次查找Middleware，效率较差；若提前将Middleware部署在路由树中性能更好，但会额外引入大量复杂代码。

### 1.4. Template页面渲染[^3]

- [x] 增加模板引擎，使用`.gohtml`文件返回渲染页面。

  [^3]: 理论上页面渲染不应该由web框架管理。

### 1.5. Multipart文件上传和下载[^4]

- [x] 上传：可重用`CopyBuffer`、设置文件权限；
- [x] 下载：使用固定header、禁用浏览器缓存、标明文件类型。

  [^4]: 文件的上传和下载一般使用oss对象存储而非服务器本身。

### 1.6. Session会话管理

- [x] 提取`Session`会话、`Store`存储、`Propagator`操作浏览器`Cookie`抽象；
- [x] 基于内存和redis的两种服务器存储实现；
- [x] 提供`SessionManager`胶水框架，暴露对外接口。

---

## 2. Go Orm框架

### 2.1. SQL语句

- [x] `QueryBuilder`作为构建SQL语句的顶级抽象，封装预编译的SQL语句和参数`Query`；
- [x] `Querier`抽象用于`SELECT`语句，`Executor`抽象用于`INSERT`、`UPDATE`、`INSERT`语句；
- [x] `Expression`是所有不同类型SQL部分拼接的标记接口；

  <details><summary>示例代码</summary>

  ```
	_ Expression = &Aggregate{}     // 内置函数
	_ Expression = &Column{}        // 列
	_ Expression = &Predicate{}     // 查询条件
	_ Expression = &RawExpr{}       // 自定义语句
	_ Expression = &value{}         // 具体值
	_ Expression = &Subquery{}      // 子查询
	_ Expression = &SubqueryExpr{}  // 子查询表达式
  ```

  </details>

- [x] 支持基础的DML语句和`WHERE`、`FROM`等关键字以及`NOT`等运算符。

### 2.2. Model元数据

- [x] 提供`registry`注册中心，复用元数据`Model`[^5]；
- [x] 对封装查询条件、提取结果集，分别提供基于反射和`unsafe`的两种实现；
- [x] 默认使用驼峰转下划线作为表名/列名，同时提供`tag`标签和`option`方法供外界修改。

  [^5]: 用户不应该动态修改元数据，在不保证并发安全的情况下提升查询效率。

### 2.3. 复杂SELECT语句

- [x] 支持聚合函数、别名、原生表达式。

  <details><summary>示例代码</summary>

  ```
  NewSelector[TestModel](db)).Select(Raw("COUNT(DISTINCT `first_name`)").As("name"), Max("Age").As("max"))
      .Where(C("Id").Eq(Raw("`age` + ?", 1).AsPredicate())
  ```

  </details>

### 2.4. UPSERT实现

- [x] 引入方言抽象`Dialect`，默认使用MySQL，暴露`option`方法供外界修改；
- [x] 替换所有反引号、提供公共方法`buildUpsert`，实现不同数据库下的`upsert`语句；
- [x] `UpsertBuilder`作为`Inserter`的中间链式调用，提供专属方法`Update`。

  <details><summary>示例代码</summary>

  ```
  NewInserter[TestModel](db).Values(&TestModel{
                  Id:        12,
                  FirstName: "Tom",
                  Age:       17,
                  LastName:  &sql.NullString{String: "Jane", Valid: true},
              }).OnDuplicateKey().ConflictColumns("Id").Update(
                  Assign("FirstName", "haha"),
                  Assign("Age", 5))
  ```

  </details>

### 2.5. Transaction事务

- [x] 抽象`Session`会话，将原有的`DB`作为实现，增加`Tx`实现，封装`DoTx`闭包方法。

  <details><summary>示例代码</summary>

  ```
  err := db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
          // do something
          return nil
      }, &sql.TxOptions{})
  ```

  </details>

### 2.6. Middleware中间件

- [x] 接入`QueryLog`记录操作日志；
- [x] 接入`OpenTelemetry`可观测性链路；
- [x] 接入`Prometheus`实现性能监控；
- [x] 接入`NoDelete`禁用删除语句；
- [x] 接入`SafeDML`禁用不使用查询条件的更新/删除；
- [x] 接入`SlowQuery`记录慢查询。

### 2.7. JOIN查询

- [x] 支持`JOIN`查询。

  <details><summary>示例代码</summary>

  ```
  t1 := TableOf(&Order{}).As("t1")
  t2 := TableOf(&OrderDetail{}).As("t2")
  t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
  t4 := TableOf(&Item{}).As("t4")
  t5 := t4.Join(t3).On(t2.C("ItemId").Eq(t4.C("Id")))
  s := NewSelector[Order](db).From(t5)
  ```

  </details>

### 2.8. AST生成模板代码

- [x] 基于AST抽象语法树与模板引擎、生成脚手架代码。

  <details><summary>示例代码</summary>

  ```
  package testdata
  
  import (
      "code-practise/orm"
      "database/sql"
  )
  
  const (
      UserName     = "Name"
      UserAge      = "Age"
      UserNickName = "NickName"
      UserPicture  = "Picture"
  )
  
  func UserNameLt(val string) orm.Predicate {
      return orm.C(UserName).Lt(val)
  }
  ...
  ```

  </details>

### 2.9. Subquery子查询

- [x] 子查询作为表达式、条件、字段、表构建SQL，支持嵌套，编写复杂查询语句时建议使用`RawQuery`。

  <details><summary>示例代码</summary>

  ```
  _ TableReference = &Subquery{}
  _ Expression     = &Subquery{}
  _ Expression     = &SubqueryExpr{}
  ```

  </details>

---

## 3. Go Cache框架

### 3.1. Memory本地缓存

- [x] 本地内存缓存，支持最大内存数[^6]、最大键值对限制；
- [x] 读时删除过期数据、开启单个goroutine随机删除部分已过期缓存。

  [^6]: 基于unsafe，无法准确获取指针引用、嵌套结构体的实际大小。

### 3.2. Redis与分布式锁[^7]

- [x] 支持redis缓存、基于Lua脚本实现分布式锁；
- [x] 支持自动续约：定时续约、锁超时续约，ctx链路超时、解锁时退出；
- [x] 支持重试策略：迭代器形态，用于用户自定义拓展和重写；
- [x] 支持`SingleFlightLock`：本地goroutine先竞争，胜者竞争分布式锁。

  [^7]: 多票表决机制的RedLock，强一致性、性能影响较大，一般不作考虑。

### 3.3. 缓存模式[^8]

- [x] 支持`ReadThrough`模式：装饰器模式重写`Get`方法，读无数据时同步/异步读数据源刷新缓存；
- [x] 支持`WriteThrough`模式：装饰器模式重写`Set`方法，写数据时同步/异步写数据源刷新缓存；
- [x] 支持`WriteBack`模式：只对缓存进行读写，用户在缓存过期时利用`OnEvicted`刷新数据源；
- [x] 支持`Refresh-Ahead`模式：需第三方中间件监听数据源数据变更事件，自动刷新缓存，如`Canal`。

  [^8]: 缓存模式无法解决一致性问题，WriteBack表现好，但安全性低。

### 3.4. 缓存三大问题与解决

- [x] `BloomFilterCache`：增强`ReadThrough`缓存，布隆过滤器，解决缓存穿透；
- [x] `SingleFlightCache`：基于`ReadThrough`缓存，读无数据时仅允许单个goroutine查询数据源，解决缓存击穿；
- [x] `RandomExpirationCache`：装饰器模式，随机缓存时间，解决缓存雪崩。

### 3.5. 缓存一致性[^9]

- [x] single_flight分段式加锁，确保单实例只有一个goroutine参与分布式锁竞争[^10]；

  [^9]: 一致性问题的根源：并发更新(分布式锁)、部分失败(分布式事务)，一致性问题本质上无解。
  [^10]: 另一方案为一致性哈希，确保请求打到同一台机器，仅在扩容、缩容、重启时可能存在一致性问题。

---

## 4. Go Micro框架

### 4.1. Pool连接池

- [x] 参考java中的线程池实现的连接池，支持范型、可管理任意类型连接；

  <details><summary>示例代码</summary>

  ```
  type Config struct {
      InitialCap  int                   // 初始连接数
      MaxCap      int                   // 最大连接数
      MaxIdle     int                   // 最大空闲连接数
      IdleTimeout time.Duration         // 最长空闲时间
      Factory     func() (any, error)   // 连接工厂
      Close       func(conn any) error  // 拒绝策略
  }
  ```

  </details>

- [x] `Get`：有空闲连接时获取，无空闲连接且未达到最大连接数时创建新连接，否则尝试放入请求队列；
- [x] `Put`：尝试获取请求队列，若请求队列为空则尝试归入空闲连接，失败则关闭连接。

### 4.2. RPC协议

- [x] 支持代理实现简易rpc协议，包含可携带元数据的变长请求头、可携带错误信息的变长响应头；

  <details><summary>示例代码</summary>

  ```
  type Request struct {
      // header                             // 请求头
      HeadLength uint32                     // 固定长度 head长度
      BodyLength uint32                     // 固定长度 body长度
      MessageId  uint32                     // 固定长度 消息id
      Version    uint8                      // 固定长度 版本号
      Compressor uint8                      // 固定长度 压缩算法
      Serializer uint8                      // 固定长度 序列化协议
  
      ServiceName string                    // 可变长度 服务名
      MethodName  string                    // 可变长度 方法名
  
      Meta map[string]string                // 可变长度 元数据
  
      // Data use any can not confirm type  // 请求体
      Data []byte                           // 可变长度 请求体
  }

  type Response struct {
      // header                             // 请求头
      HeadLength uint32                     // 固定长度 head长度
      BodyLength uint32                     // 固定长度 body长度
      MessageId  uint32                     // 固定长度 消息id
      Version    uint8                      // 固定长度 版本号
      Compressor uint8                      // 固定长度 压缩算法
      Serializer uint8                      // 固定长度 序列化协议
  
      Error []byte                          // 可变长度 error
  
      // Data use any can not confirm type  // 请求体
      Data []byte                           // 可变长度 请求体
  }
  
  ```

  </details>

- [x] 支持json、proto以及用户自定义序列化协议；
- [x] 支持gzip以及用户自定义压缩算法，默认不压缩；
- [x] `client`使用连接池，用于管理和复用tcp连接；
- [x] 支持one-way调用，服务端异步处理直接返回[^11]；
- [x] 支持链路超时控制，重建`context.Context`，中断由业务本身处理。

  [^11]: 没有必要支持异步、回调，用户可自行利用goroutine、sync包实现。

### 4.3. 注册中心

- [x] 支持服务注册：基于`etcd`租约API实现注册中心，服务启动时机由用户调用`Start`方法决定；

  <details><summary>示例代码</summary>

  ```  
  type Registry interface {
      Register(ctx context.Context, si ServiceInstance) error
      UnRegister(ctx context.Context, si ServiceInstance) error
      ListServices(ctx context.Context, name string) ([]ServiceInstance, error)
      Subscribe(name string) (<-chan Event, error)
      io.Closer
  }
  ```

  </details>

- [x] 支持服务发现：基于`grpc.resolver`，启动时从注册中心拉取全部服务信息，持续监听etcd中服务变化、全量更新服务信息[^12]；
- [x] 服务端-注册中心连接失败：连接基于服务端主动续约，受租约长短、重试机制影响，重试失败后注册中心通知客户端；
- [x] 客户端-注册中心连接失败：解决方案为停止服务(CP)或使用本地缓存数据连接服务端(AP)；
- [x] 客户端-服务端连接失败：暂时将服务端从节点列表剔除，后续考虑恢复。

  [^12]: 无法获取容器地址：只能考虑使用容器专属注册与发现方式、容器使用宿主机网络或启动时将宿主机IP作为环境变量注入。

### 4.4. 负载均衡、路由和集群[^13]

- [x] 支持静态负载均衡策略：轮询、加权(平滑)轮询、随机、加权随机、哈希[^14]；
- [x] 支持动态负载均衡策略：最快响应时间、最少连接数、最少请求数；
- [x] 支持路由策略：服务端在注册中心额外注册分组信息，实现分组路由，路由会影响负载均衡的准确性；
- [x] 支持cluster集群：解决换节点重试(轮询或grpc重试)、快速失败、广播、组播功能。

  [^13]: 不借助网关或第三方中间件的客户端负载均衡策略并不准确，理论上发起调用后的结果需要反馈给服务发现组件、cluster组件、路由组件和负载均衡组件(大多数微服务框架并不支持)。
  [^14]: 一致性哈希存在构造复杂、服务上下线难以处理等问题，更适合在网关层面实现。

### 4.5. 可用性与可观测性[^15]

- [x] 支持静态故障检测算法：令牌桶、漏桶、固定窗口和滑动窗口[^16]；
- [x] 支持基于本地内存的单机限流、基于redis的集群限流；
- [x] 限流请求只走快路径/简易路径(基于redis的滑动窗口算法)[^17]；
- [x] 接入`Metrics`，记录请求数、错误数、响应时间；
- [x] 接入`OpenTelemetry`可观测性链路。

  [^15]: 熔断、限流、降级本质上都是故障处理策略(故障检测-故障处理-故障恢复)。
  [^16]: 动态故障检测算法包括根据服务状态动态判断，如错误率、响应时间、或BBR算法，一般不需要使用。
  [^17]: 拒绝策略与业务强相关，包括快路径、返回固定响应、缓存请求、转为异步、转发到其他服务器(较少见)。
