# Code Practise

---

## 1. Go Web框架

### 1.1. Router路由

- [x] 抽象`Server`接口，支持`HTTPServer`实现；
- [x] 以`/`作为分割符构造路由树，支持静态匹配[^1]；
- [x] 支持不同节点类型，实现高级路由，包括`/*`通配符匹配、`/:id`路径参数。

  [^1]: Gin框架使用了前缀树，查找速度快；但代码过于复杂、因此不考虑。

### 1.2. Context上下文

- [x] 封装`*http.Request`与`http.ResponseWriter`，用于Middleware；
- [x] 存储url参数、路径参数、请求体、表单参数；
- [x] 封装`StringValue`，支持返回值类型转化语法糖。

---

  <details><summary>示例代码</summary>

  ```
    req, err := ctx.QueryValue("file").String()
  ```

  </details>

---

### 1.3. Middleware中间件

- [x] 重写路由查找、支持节点级别的Middleware[^2]；
- [x] `AccessLog`记录请求日志；
- [x] `OpenTelemetry`接入可观测性链路；
- [x] `Prometheus`实现性能监控；
- [x] `Errhandle`返回错误页面；
- [x] `Recover`支持从错误中恢复。

  [^2]: 实现方法次啊用匹配路由后二次查找Middleware，效率较差；提前将Middleware部署在路由树中性能更好，但会引入大量复杂代码。

### 1.4. Template页面渲染[^3]

- [x] 增加模板引擎，使用`.gohtml`文件返回渲染页面。

  [^3]: 仅用于代码练习，理论上页面渲染不应该由web框架管理。

### 1.5. Multipart文件上传和下载[^4]

- [x] 上传：可重用`CopyBuffer`、设置文件权限；
- [x] 下载：使用固定header、禁用浏览器缓存、标明文件类型。

  [^4]: 仅用于代码练习，文件的上传和下载一般使用oss对象存储而非服务器本身。

### 1.6. Session会话管理

- [x] 提取`Session`会话、`Store`存储、`Propagator`操作浏览器`Cookie`抽象；
- [x] 基于内存和`redis`的两种服务器存储实现；
- [x] `SessionManager`胶水框架，暴露对外接口。

---

## 2. Go Orm框架

### 2.1. SQL语句

- [x] `QueryBuilder`作为构建SQL语句的顶级抽象，封装预编译的SQL语句和参数`Query`；
- [x] `Querier`抽象用于`SELECT`语句，`Executor`抽象用于`INSERT`、`UPDATE`、`INSERT`语句；
- [x] `Expression`是所有不同类型SQL部分拼接的标记接口；

---

  <details><summary>示例代码</summary>

  ```
	_ Expression = &Aggregate{}     // 内置函数
	_ Expression = &Column{}        // 列名
	_ Expression = &Predicate{}     // 查询条件
	_ Expression = &RawExpr{}       // 自定义语句
	_ Expression = &value{}         // 具体值
	_ Expression = &Subquery{}      // 子查询
	_ Expression = &SubqueryExpr{}  // 子查询
  ```

  </details>

---

- [x] 支持基础的DML语句和`WHERE`、`FROM`等关键字以及`NOT`等运算符。

### 2.2. Model元数据

- [x] 提供`registry`注册中心，复用元数据`Model`[^5]；
- [x] 对封装查询条件、提取结果集，分别提供基于反射和`unsafe`的两种实现；
- [x] 默认使用驼峰转下划线作为表名/列名，同时提供`tag`标签和`option`方法供外界修改。

  [^5]: 用户不应该动态修改元数据，在不保证并发安全的情况下提升查询效率。

### 2.3. 复杂SELECT语句

- [x] 支持聚合函数、别名、原生表达式。

---

  <details><summary>示例代码</summary>

  ```
  NewSelector[TestModel](db)).Select(Raw("COUNT(DISTINCT `first_name`)").As("name"), Max("Age").As("max"))
      .Where(C("Id").Eq(Raw("`age` + ?", 1).AsPredicate())
  ```

  </details>

---

### 2.4. UPSERT实现

- [x] 引入方言抽象`Dialect`，默认使用MySQL，暴露`option`方法供外界修改；
- [x] 替换所有反引号、提供公共方法`buildUpsert`，实现不同数据库下的`upsert`语句；
- [x] `UpsertBuilder`作为`Inserter`的中间链式调用，提供专属方法`Update`。

---

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

---

### 2.5. Transaction事务

- [x] 抽象`Session`会话，将原有的`DB`作为实现，增加`Tx`实现，封装`DoTx`闭包方法。

---

  <details><summary>示例代码</summary>

  ```
  err := db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
          // do something
          return nil
      }, &sql.TxOptions{})
  ```

  </details>

---

### 2.6. Middleware中间件

- [x] `QueryLog`记录操作日志；
- [x] `OpenTelemetry`接入可观测性链路；
- [x] `Prometheus`实现性能监控；
- [x] `NoDelete`禁用删除语句；
- [x] `SafeDML`禁用不使用查询条件的更新/删除；
- [x] `SlowQuery`记录慢查询。

### 2.7. JOIN查询

- [x] 支持`JOIN`查询。

### 2.8. AST抽象语法树与模板代码

- [x] 使用AST生成脚手架代码。

### 2.9. Subquery子查询

- [x] 实现基础子查询，复杂查询应使用`RawQuery`。

---

## 3. Go Cache框架

### 3.1. Memory本地缓存

### 3.2. Redis缓存

### 3.3. 缓存容量限制

### 3.4. 一致性问题与解决

---

## 4. Go Micro框架

### 4.1. Pool连接池

### 4.2. RPC协议设计实现

### 4.3. Etcd注册与发现

### 4.4. 负载均衡

### 4.5. Cluster集群

### 4.6. 可用性与可观测性

