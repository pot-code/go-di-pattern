# 使用

## DI

### 声明组件

一个典型的组件声明：

```go
type ExampleService struct {
	RedisClient *db.RedisDB `dep:"RedisDB"`
}

func (es ExampleService) Constructor() *ExampleService {
	return &ExampleService{es.RedisClient}
}
```

- 结构成员 tag 定义的 `dep` 字段用于标志需要注入的依赖组件的类型名，字段类型必须为指针类型
- 每个组件结构都要包含一个 `Constructor()` method，而且 receiver 必须是值类型，**不能是指针类型**
- `Constructor()` 返回值是当前组件的指针类型，该方法通过 DI 容器调用，**切勿自主调用**。可以认为此 method 调用时，receiver
  里标记了 `dep` 的字段已经被注入了依赖，将这些字段传入要返回的结构体中，其他字段根据需要进行初始化

对于一个没有依赖的组件（依赖树的叶子节点）：

```go
type RedisDB struct {
	*redis.Client
}

func (rdb RedisDB) Constructor() *RedisDB {
	return &RedisDB{newRedisClient()}
}

func newRedisClient() *redis.Client {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	client := redis.NewClient(&redis.Options{
		DB: 0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	return client
}
```

同样需要一个 `Constructor()` method，在里面放入初始化代码即可

### 初始化 DI 容器

```go
dic := container.NewDIContainer()
```

### 注册组件

项目在每个 package 下放了一个 `init.go` 文件，里面包含两个函数：`Injection()` 和 `InitPackage()`，`Injection` 用于注册当前 package 下的组件：

```go
func Injection(dic *container.DIContainer) {
	dic.Register(new(LoginController))
	dic.Register(new(JWTMiddleware))
}
```

只需要传入结构体默认值的指针即可，注册默认以结构体定义的类型名作为组件名。

`Register()` 还可以接受额外的 namespace 参数， 用于指定当前组件隶属的命名空间。可以有多个命名空间前缀，按照传参顺序拼接，以 `:` 作为拼接符：

```go
dic.Register(new(LoginService), "root", "sub", "subsub")
dic.Get("root:sub:subsub:LoginService")

// dep tag
type ExampleController struct {
	*service.LoginService `dep:"root:sub:subsub:LoginService"`
}
```

### 使用组件

>组件一定要先注册再使用，并且等待全部组件注册完毕再使用，否则可能出现依赖没有注册的情况，这样就无法构建完整的组件

```go
func InitPackage(dic *container.DIContainer) {
	lc := dic.Get("LoginController").(*LoginController)
	jm := dic.Get("JWTMiddleware").(*JWTMiddleware)

	homeMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware, jm.ValidateMiddleware, jm.RefreshTokenMiddleware)
	loginMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware)
	logoutMiddleware := Compose(ErrorHandlingMiddleware, jm.ValidateMiddleware)

	http.HandleFunc("/login", loginMiddleware(lc.handleLogin))
	http.HandleFunc("/logout", logoutMiddleware(lc.handleLogout))
	http.HandleFunc("/home", homeMiddleware(lc.handleHome))
}
```

### 接口

接口注入也是支持的：

```go
type ILowLevel interface {
	Counter() int
}

type TopLevelComponent struct {
	ILowLevel `dep:"LowLevelComponent"`
}
```

### 循环依赖

DI 包含循环依赖的检测代码，根据错误提示操作

## 运行

`go run ./`