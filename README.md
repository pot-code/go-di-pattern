# 使用

## DI

### 声明组件

一个典型的组件声明：

```go
type ExampleService struct {
	RedisClient *db.RedisDB `dep:""`
}

func (es ExampleService) Constructor() *ExampleService {
	return &ExampleService{es.RedisClient}
}
```

- 结构成员 tag 定义的 `dep` 键用于标记需要注入的字段，字段类型必须为指针类型或接口类型。不要使用 embedded field，会导致接口注入失败
- 每个组件结构都要包含一个 `Constructor()` method，而且 receiver **必须是值类型，不能是指针类型**
- `Constructor()` 返回值是当前组件的指针类型，该方法通过 DI 容器调用，**不建议自主调用**（需要手动获取组件时除外）。可以认为此 method 调用时，receiver
  里标记了 `dep` 的字段已经被注入了依赖，将这些字段传入要返回的结构体中，其他字段根据需要进行初始化

`dep` 的值为空串时，在已注册的组件中选出一个符合条件的赋给这个字段。需要自主指定时，需要提供目标类型的完整类型名，即包括模块名、包名；如果手动指定的类型不匹配当前字段类型会报错。

embedded field 会让当前结构体实现要注入的接口类型，形成自依赖关系。尽量使用独立的成员变量名

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

调用 `dic.Register(new(类型名))`，注册组件：

```go
dic.Register(new(RedisDB))
```

当前组件的依赖也会自动进行注册，所以可以只注册顶层组件和最底层（没有其他依赖项）的组件就能构建完整的依赖关系，前提是要有定义完善的 `Constructor` method

### 获取组件

>组件一定要先注册再使用，并且等待全部组件注册完毕再使用，否则可能出现依赖没有注册的情况，这样就无法构建完整的组件

需要使用组件类型的完整路径获取组件指针，例如：

```go
dic.Get("github.com/pot-code/go-di-pattern/route/LoginController")
```

这样设计是为了避免命名冲突，可以在注册时使用一个变量保存引用，然后通过这个引用获取：

```go
instance := new(route.LoginController)
dic.Register(instance)
dic.Populate() // 调用会触发初始化

instance.DoXxx()
```

### 接口

接口注入也是支持的：

```go
type ILowLevel interface {
	Counter() int
}

type TopLevelComponent struct {
	LowLevel ILowLevel `dep:""`
}
```

详见 `container/di_test.go` 里的示例

### 循环依赖

DI 包含循环依赖的检测代码，根据错误提示操作

## 运行

先启动本地 Redis，再执行 `go run ./`

### 路由

- `/login`，提供 form 字段 `username`，获取 JWT
- `/home`，访问主页，需要 JWT 验证
- `/logout`，登出，注销当前的 JWT