## rcd
### 逻辑过程：
1. udesk工单满足udesk触发器发布文档的功能条件时，udesk实时向s1mple发起POST请求
2. Post请求中携带工单中所需要的数据，比如工单中的cloudId、jira、标题、描述等
3. s1mple将来自udesk请求中的body json，序列化为Document结构体
4. 初始化Document结构体时，与配置中给定的username确定当前具体发布者
5. 根据doc_go_template中需要的字段，将Document结构体渲染进去，返回故障文档的html内容
6. 在故障文档的html内容中获取图片文件，依次创建Img结构体，并传入到Document的Imgs channel中
7. 对故障文档html内容进行修饰，删除多余的html标签，删除多余的字符串
8. 构造发布到confluence文档的Post请求所需要的body json，其中，将最终修饰后的故障文档html内容传入
9. 向confluence发起发布文档的Post请求创建文档
10. 文档创建成功后将返回
11. 从Document的Imgs channel中获取Img结构体的数量，运行其数量的goroutine，并发下载图片和上传图片