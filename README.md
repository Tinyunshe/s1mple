# rcd

### 功能更新记录：

~~1、程序日志需要加~~

~~2、在发布文档之前判断文档是否存在（发布前校验）~~

~~3、发布文档的内容格式~~

~~4、发布文档中的图片需要处理~~

~~5、如果confluence异常，发布请求时的超时设置，重试设置，错误处理，可以考虑将错误发送到企微，需要加~~

~~6、考虑是否要发布到空间主页面（比如wushuting的主页面），还是就在顶级页面~~

~~7、服务端鉴权~~

~~8、根据存放img目录的大小自动回收img文件防止越来越大~~

~~9、img download和upload的并发逻辑需要优化，存在图片还没下载好，上传逻辑就开始，导致的一系列问题~~

~~0、工单描述的图片和附件需要处理~~

~~11、附件是非图片格式的不进行下载和上传~~

~~12、故障文档标题格式修改，举例：容器平台-网络-问题标题-cloudId~~

~~13、从工单系统获取 宏 ，删除工单回复中有 宏 的内容~~

~~14、将工单回复顺序反转~~

~~15、去掉您好，你好，等开头字符~~

~~16、发现文档中原始大小有点大，可以考虑将img大小调整到400 <ac:image ac:height="400"><ri:attachment ri:filename="mceclip3_1710839260755_l7lsg.png" /></ac:image>~~

~~17、重新组织adorn架构，将所有修饰字段的操作归纳到该package中~~

14、将工单回复顺序反转

15、去掉您好，你好，等开头字符

16、发现文档中原始大小有点大，可以考虑将img大小调整到400 <ac:image ac:height="400"><ri:attachment ri:filename="mceclip3_1710839260755_l7lsg.png" /></ac:image> 

# Notify

### Review

1、组织架构

编译：

docker build -t tinyunshe/s1mple:$(date +%s) .   && docker image ls

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o s1mple