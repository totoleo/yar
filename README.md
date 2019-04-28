#### Yar.go
-----
[Yar.go](https://github.com/totoleo/yar) 是一个[YAR，一个轻量级的跨语言RPC框架](https://github.com/laruence/yar)的客户端实现

#### 特性

1. 支持http客户端
2. 支持动态参数列表，但不支持默认值

-----

#### Example Client

```go
package main

import (
    "context"
	"fmt"
	
	"github.com/totoleo/yar"
	"github.com/totoleo/yar/client"
)

func main() {
    
    //初始化一个客户端。
    //目前仅支持http,https 的Yar服务端
	c, err := client.NewClient("http://127.0.0.1:8080")
    
	if err != nil {
		fmt.Println("error", err)
	}

	//这是默认值
    //定义Yar的服务端方法返回值
	var ret interface{}

	callErr := c.Call(context.TODO(),"echo", &ret)
    
    //错误判断
	if callErr != nil {
		fmt.Println("error", callErr)
	}
    
	fmt.Println("data", ret)
}

```



