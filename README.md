1、程序日志需要加

2、在发布文档之前判断文档是否存在（发布前校验）

3、发布文档的内容格式

4、发布文档中的图片需要处理

5、如果confluence异常，发布请求时的超时设置，重试设置，需要加

6、考虑是否要发布到空间主页面（比如wushuting的主页面），还是就在顶级页面

7、~~服务端鉴权~~

编译：

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o s1mple