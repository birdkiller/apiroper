# apiroper
A package which offers abilities to make apis called togather, no matter synchronous or asynchronous.

# package apiroper 提供请求聚合功能
## V0.0.2

## 功能说明：
. 通过读取模板，将多个请求聚合成一个目标输出，根据依赖关系自动选择同步或异步请求并根据模版聚合输出。

## 用法：
1. 加载：
  `import "saas/common/utils/apiroper"
			templates := map[string]apiroper.Template{}
			apiroper.Load(templates)`
      
2. 调用：
  `args := map[string]interface{}{
			"mobile": "13655566676",
			}
	    	output, err := apiroper.Call("getUserInfo", args)`
        
3. 名词解释&模版规则：
  + Template - api集合模板，代表一个由资源api组成的调用集合。
    + Template.Output - 调用集合的输出key-value键值对，其中key必须为字符串、
				值可以是任何类型。如果值包含<<变量路径>>时，<<变量路径>>只能整体替换普
				通变量的全部，或者整体替换map、slice、object的值部分。
    + 例如：
					{"v1":"<<input.argv1>>", "v2":2}	正确
					{"<<input.argv1>>":"1", "v2":2}		错误
					{"v1":"<<input.argv1>>hello"}		错误
    + Template.Resources - 组成该api集合模版的key-value键值对，其中key为
				api资源注册标识，value为api资源对象。当调用程序通过api资源获得具体资源
				内容时，会把获得的资源内容存放在内存map中该key对应的值中，以供调用方以
				<<变量路径>>的形式查找和使用。
  + Resource - api资源，代表一个单一资源api(目前仅支持http请求)。
    + Resource.Url - 请求URL，当包含<<变量路径>>时，可以替换URL中的部分内容。
    + 例如：
					Url:`http://127.0.0.1:8080/account/login`
					Url:`http://127.0.0.1:8080/im/getUserInfo?token=<<login.data.Token>>`
    + Resource.Method - http请求method，目前仅支持POST
    + Resource.Input - api资源请求参数，目前仅支持json格式。key-value键
        值对，其中key为字符串，value可以使任意类型。如果value包含<<变量路径>>
				时，<<变量路径>>只能整体替换普通变量的全部，或者整体替换map、slice、
				object的值部分。
    + Resource.Header - api资源请求头，key-value键值对，其中key为字符串，
				value为字符串值，value也可以由<<变量路径>>整体替换。
    + 例如：
					Header:{"Content-Type":"application/json", "Token":"<<login.data.Token>>"}
    + Resource.Timeout - api资源获取超时时间秒数、暂未实现
  + 变量路径 - 一个由`<<`和`>>`包围的字符串组成的标签，一个变量路径代表一个
			由调用输入或请求获得的值。形如`<<login.data.SessionToken>>`，代表内存map
			中以["login"]["data"]["SessionToken"]取得的值。而`<<login.data.Users.0.Id>>`,
			等同于`<<login.data.Users[0].Id>>`，代表内存map中以["login"]["data"]["Users"][0]["Id"]
			取得的值，该路径大小写敏感。
			变量路径可以出现在Template.Output、Resource.Input、Resource.Url、
			Resource.Header中，具体规则请参考以上对象说明。
  + 路径节点 - 一个由不包含`.`、`<`、`>`、`[`、`]`字符组成的字符串，表示一个
			map变量中的key或者slice变量中的下标或者一个最终变量的名称,形如`login`。
			路径节点通过`.`或`[整数]`连接组成变量路径，其中`.`可连接map或slice类型的
			变量，`[整数]`可连接slice类型的变量，连接map类型变量时，节点值不允许为数值类型，
			如`0`、`1`、`2`...，而连接slice类型变量时，节点值必须为数值类型。
      
4. 备注：
  + 其中templates可由固定协议文本导入或代码直接编写。
  + apiroper/test包内有完整测试用例可供参考
  
  
## TODO
1. 支持多种协议或进行接口化（目前只支持http.POST）
2. 支持子请求timeout配置
3. 提供全局事务一致性支持
    
