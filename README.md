# Code Practise

## 1. Go Web框架

### 1.1. Router路由

- [x] 抽象`Server`接口，支持`HTTPServer`实现，使用`Option`模式；
- [x] 以`/`作为分割符构造路由树，支持静态匹配，校验非法参数`//`；
- [x] 支持不同节点类型，实现高级路由，包括`/*`通配符匹配、`/:id`路径参数。

### 1.2. Context上下文

- [x] 封装请求与响应，禁止直接修改`*http.Request`与`http.ResponseWriter`；
- [x] 存储查询参数`?`、路径参数、请求体、表单，封装`StringValue`。

### 1.3. Middleware中间件

- [x] 支持节点级别的Middleware；
- [x] `AccessLog`请求日志；
- [x] `OpenTelemetry`可观测性链路；
- [x] `Prometheus`性能监控；
- [x] `Errhandle`返回错误页面；
- [x] `Recover`从错误中恢复。

### 1.4. Template页面渲染[^1]

- [x] 使用`.gohtml`文件返回渲染页面。

  [^1]: 页面渲染不应该由web框架管理，此处仅用于代码练习。

### 1.5. Multipart文件上传和下载[^2]

- [x] 上传：可重用`CopyBuffer`、设置文件权限；
- [x] 下载：使用固定header、禁用浏览器缓存、标明文件类型。

  [^2]: 文件的上传和下载一般使用oss对象存储，此处仅用于代码练习。

### 1.6. Session会话管理

- [x] 提取`Session`会话、`Store`存储、`Propagator`操作浏览器`Cookie`抽象；
- [x] 提供基于内存和`redis`的两种存储实现、提供刷新功能；
- [x] `SessionManager`胶水框架，暴露对外接口。

## 2. Go Orm框架

## 3. Go Cache框架

## 4. Go Micro框架
