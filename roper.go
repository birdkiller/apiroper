/*
	package apiroper 提供请求聚合功能
	版本：V0.0.2
	功能说明：
		通过读取模板，将多个请求聚合成一个目标输出，根据依赖关系
	自动选择同步或异步请求并根据模版聚合输出。
	用法：
		1、加载：
			import "saas/common/utils/apiroper"
			templates := map[string]apiroper.Template{}
			apiroper.Load(templates)
		2、调用：
			args := map[string]interface{}{
			"mobile": "13655566676",
			}
	    	output, err := apiroper.Call("getUserInfo", args)
		3、名词解释&模版规则：
			Template - api集合模板，代表一个由资源api组成的调用集合。
				Template.Output - 调用集合的输出key-value键值对，其中key必须为字符串、
				值可以是任何类型。如果值包含<<变量路径>>时，<<变量路径>>只能整体替换普
				通变量的全部，或者整体替换map、slice、object的值部分。
				例如：
					{"v1":"<<input.argv1>>", "v2":2}	正确
					{"<<input.argv1>>":"1", "v2":2}		错误
					{"v1":"<<input.argv1>>hello"}		错误
				Template.Resources - 组成该api集合模版的key-value键值对，其中key为
				api资源注册标识，value为api资源对象。当调用程序通过api资源获得具体资源
				内容时，会把获得的资源内容存放在内存map中该key对应的值中，以供调用方以
				<<变量路径>>的形式查找和使用。

			Resource - api资源，代表一个单一资源api(目前仅支持http请求)。
				Resource.Url - 请求URL，当包含<<变量路径>>时，可以替换URL中的部分内容。
				例如：
					Url:`http://127.0.0.1:8080/account/login`
					Url:`http://127.0.0.1:8080/im/getUserInfo?token=<<login.data.Token>>`
				Resource.Method - http请求method，目前仅支持POST
				Resource.Input - api资源请求参数，目前仅支持json格式。key-value键
				值对，其中key为字符串，value可以使任意类型。如果value包含<<变量路径>>
				时，<<变量路径>>只能整体替换普通变量的全部，或者整体替换map、slice、
				object的值部分。
				Resource.Header - api资源请求头，key-value键值对，其中key为字符串，
				value为字符串值，value也可以由<<变量路径>>整体替换。
				例如：
					Header:{"Content-Type":"application/json", "Token":"<<login.data.Token>>"}
				Resource.Timeout - api资源获取超时时间毫秒数

			变量路径 - 一个由`<<`和`>>`包围的字符串组成的标签，一个变量路径代表一个
			由调用输入或请求获得的值。形如`<<login.data.SessionToken>>`，代表内存map
			中以["login"]["data"]["SessionToken"]取得的值。而`<<login.data.Users.0.Id>>`,
			等同于`<<login.data.Users[0].Id>>`，代表内存map中以["login"]["data"]["Users"][0]["Id"]
			取得的值，该路径大小写敏感。
			变量路径可以出现在Template.Output、Resource.Input、Resource.Url、
			Resource.Header中，具体规则请参考以上对象说明。

			路径节点 - 一个由不包含`.`、`<`、`>`、`[`、`]`字符组成的字符串，表示一个
			map变量中的key或者slice变量中的下标或者一个最终变量的名称,形如`login`。
			路径节点通过`.`或`[整数]`连接组成变量路径，其中`.`可连接map或slice类型的
			变量，`[整数]`可连接slice类型的变量，连接map类型变量时，节点值不允许为数值类型，
			如`0`、`1`、`2`...，而连接slice类型变量时，节点值必须为数值类型。
        4、备注：
			其中templates可由固定协议文本导入或代码直接编写。
			apiroper.main包内有完整测试用例可供参考
	TODO:
		1、支持多种协议或进行接口化（目前只支持http.POST）
		2、支持子请求timeout配置
		3、提供全局事务一致性支持
*/
package apiroper

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var callstacks map[string]*sync.Pool

func init() {
	callstacks = map[string]*sync.Pool{}
}

// ResourceCaller 资源请求
type ResourceCaller struct {
	Resource
	Id              string
	stack           *CallStack
	urlarguments    map[string]*argument
	headerarguments map[string]*argument
	inputarguments  map[string]*argument
	dependences     []string
	mutex           *sync.Mutex
	iscalled        bool
	//Supplies    []string
}

// CallStack 调用序列
type CallStack struct {
	Output       map[string]interface{}
	dependences  []string
	requirements map[string]*argument
	context      map[string]interface{}
	resources    map[string]*ResourceCaller
}

// AddResources 新增资源
func (self *CallStack) addResources(name string, res *ResourceCaller) {
	res.stack = self
	self.resources[name] = res
	return
}

// GetResource 获得一个资源的对象
func (self *CallStack) getResource(name string) *ResourceCaller {
	res, ok := self.resources[name]
	if ok {
		return res
	} else {
		return nil
	}
}

// Setup 使调用序列生效
func (self *CallStack) setup() {
	// 调用序列
	arguments := getarguments(self.Output)
	for _, arg := range arguments {
		self.requirements[arg.idkey] = arg
	}
	// 依赖
	self.dependences = []string{}
	for _, req := range self.requirements {
		if req.base != "" {
			self.dependences = append(self.dependences, req.base)
		}
	}
	for _, res := range self.resources {
		// 分析url
		urlargs := getarguments(res.Url)
		for _, arg := range urlargs {
			res.urlarguments[arg.idkey] = arg
		}
		// 分析header
		headerargs := getarguments(res.Header)
		for _, arg := range headerargs {
			res.headerarguments[arg.idkey] = arg
		}
		// 分析入参
		inputargs := getarguments(res.Input)
		for _, arg := range inputargs {
			res.inputarguments[arg.idkey] = arg
		}
		// 依赖
		res.dependences = []string{}
		reqmap := unionmap(res.urlarguments, res.headerarguments, res.inputarguments)
		for _, req := range reqmap {
			if req.base != "" {
				res.dependences = append(res.dependences, req.base)
			}
		}
	}
	return
}

func (self *CallStack) getOutput(data interface{}) interface{} {
	datastring, ok := data.(string)
	if ok {
		//字符串处理逻辑：直接在输入参数里搜索是否存在对应的key，若存在则取值，
		//否则即为固定字符串，直接输出
		arg, _ := self.requirements[datastring]
		if arg != nil {
			v, _ := arg.getValue(self.context)
			return v
		} else {
			return datastring
		}
	}

	dataslice, ok := data.([]interface{})
	if ok {
		outslice := []interface{}{}
		for _, d := range dataslice {
			outslice = append(outslice, self.getOutput(d))

		}
		return outslice
	}

	datamap, ok := data.(map[string]interface{})
	if ok {
		outmap := map[string]interface{}{}
		for k, d := range datamap {
			outmap[k] = self.getOutput(d)

		}
		return outmap
	}

	return data
}

// call 调用
func (self *CallStack) call(args map[string]interface{}) (out interface{}, err error) {
	finishgroup := sync.WaitGroup{}
	for _, dep := range self.dependences {
		res, _ := self.resources[dep]
		if res == nil {
			return nil, fmt.Errorf("Unsupported resource %s", dep)
		}
		go res.call(&finishgroup)
		finishgroup.Add(1)
	}
	// 等待全部执行完
	finishgroup.Wait()

	// 根据输出模版获取输出
	out = self.getOutput(self.Output)
	return
}

func (self *CallStack) reset() {
	self.context = map[string]interface{}{}
}

// call
func (self *ResourceCaller) call(g *sync.WaitGroup) error {
	defer g.Done()
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.iscalled != false {
		return nil
	}

	// 调用标记防止重复调用
	self.iscalled = true

	// 先调用依赖资源
	finishgroup := sync.WaitGroup{}
	for _, dep := range self.dependences {
		res, _ := self.stack.resources[dep]
		if res == nil {
			continue
		}
		go res.call(&finishgroup)
		finishgroup.Add(1)
	}

	// 等待全部执行完毕
	finishgroup.Wait()

	url := self.getUrl()
	headers := self.getHeader()
	input := self.getInput(self.Input)
	data, _ := json.Marshal(input)
	complete := make(chan error) // 结果chan
	if self.Timeout == 0 {
		self.Timeout = DEFAULT_TIMEOUT_DURATION
	}
	timeout := time.After(time.Duration(self.Timeout) * time.Millisecond) // 超时chan
	go func() {
		code, body, err := post(url, data, headers)

		if code == 200 {
			ret := map[string]interface{}{}
			json.Unmarshal(body, &ret)
			self.stack.context[self.Id] = ret
		}

		complete <- err

	}()

	select {
	case err := <-complete:
		return err
	case <-timeout:
		err := fmt.Errorf("[%s] calls timeout", self.Url)
		log.Println(err)
		return err
	}

}

func (self *ResourceCaller) getUrl() string {
	url := self.Url
	for k, arg := range self.urlarguments {
		v, err := arg.getValue(self.stack.context)
		if err != nil {
			return url
		}
		vs := fmt.Sprintf("%v", v)
		url = strings.Replace(url, k, vs, -1)
	}
	return url
}

func (self *ResourceCaller) getHeader() map[string]string {
	header := self.Header
	for hk, hv := range header {
		arg, _ := self.headerarguments[hv]
		if arg != nil {
			v, _ := arg.getValue(self.stack.context)
			header[hk] = fmt.Sprintf("%v", v)

		}
	}
	return header
}

func (self *ResourceCaller) getInput(data interface{}) interface{} {
	datastring, ok := data.(string)
	if ok {
		//字符串处理逻辑：直接在输入参数里搜索是否存在对应的key，若存在则取值，
		//否则即为固定字符串，直接输出
		arg, _ := self.inputarguments[datastring]
		if arg != nil {
			v, _ := arg.getValue(self.stack.context)
			return v
		} else {
			return datastring
		}
	}

	dataslice, ok := data.([]interface{})
	if ok {
		inputslice := []interface{}{}
		for _, d := range dataslice {
			inputslice = append(inputslice, self.getInput(d))

		}
		return inputslice
	}

	datamap, ok := data.(map[string]interface{})
	if ok {
		inputmap := map[string]interface{}{}
		for k, d := range datamap {
			inputmap[k] = self.getInput(d)

		}
		return inputmap
	}

	return data
}

// 根据data获取参数列表
func getarguments(data interface{}) []*argument {
	args := []*argument{}
	datastring, ok := data.(string)
	if ok {
		args = analyze(datastring)
		return args
	}

	dataslice, ok := data.([]interface{})
	if ok {
		for _, d := range dataslice {
			subargs := getarguments(d)
			for _, subarg := range subargs {
				args = append(args, subarg)
			}
		}
		return args
	}

	datamap, ok := data.(map[string]interface{})
	if ok {
		for _, d := range datamap {
			subargs := getarguments(d)
			for _, subarg := range subargs {
				args = append(args, subarg)
			}
		}
		return args
	}
	return args
}

// 加载模板
func Load(templates map[string]Template) {
	for name, template := range templates {
		pool := new(sync.Pool)
		pool.New = func() interface{} {
			stack := &CallStack{
				Output:       template.Output,
				dependences:  []string{},
				requirements: map[string]*argument{},
				context:      map[string]interface{}{},
				resources:    map[string]*ResourceCaller{},
			}
			for rname, res := range template.Resources {
				rescaller := &ResourceCaller{
					Id:              rname,
					Resource:        res,
					urlarguments:    map[string]*argument{},
					headerarguments: map[string]*argument{},
					inputarguments:  map[string]*argument{},
					mutex:           new(sync.Mutex)}
				stack.addResources(rname, rescaller)
			}
			stack.setup()
			return stack
		}

		callstacks[name] = pool
	}
	return
}

// Call
func Call(key string, args map[string]interface{}) (out interface{}, err error) {
	pool, _ := callstacks[key]
	if pool == nil {
		return nil, fmt.Errorf("Unsupported key")
	}
	callstack := pool.Get().(*CallStack)
	defer func() {
		callstack.reset()
		pool.Put(callstack)
	}()

	for k, v := range args {
		callstack.context[k] = v
	}
	out, err = callstack.call(args)

	return
}
