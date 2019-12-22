# Godict

近期一直再使用golang语言开发一些工具，相关的后端技术链(golang+orm+postgresql+gin+jwt+logrus)和对应前端的技术链(vue+iview+axios+vue-router)基本已经打通了，[项目地址]( https://github.com/qiuzhiqian/etc_tsp )。但是想到除了这一套前后端的东西外，命令行的一些操作也是不可避免的。因此就找到了cobra这个应用广泛的第三方命令行库，并借这个小项目练一下手。

## 功能

golang天然的带有网络操作的优势，所以直接借用现有的第三方api服务来做一个实用的小工具。首先想到的就是词典翻译，因为这个工具我之前在学习python时就做过一个。

[python版本有道词典]( https://github.com/qiuzhiqian/Sdet )

既然需要做一个golang版本的有道词典(前期只考虑命令行)，那么第一步就需要有翻译接口。

接口获取有两种途径：

1. 模拟网页请求，然后解析html文本抓取其中的有效结果数据。这就是所谓的爬虫了。

2. 使用官方指定的api调用接口直接获取数据。  

   

第一种方法方法简单粗暴，没啥限制，但是由于时爬虫解析整个网页，如果网页结构变化了，就容易失效，而且效率也相对较低，毕竟要从一大堆数据中找出一点点有用的东西出来。

第二种方法是官方提供的接口，所以基本是长期有效的，相对很稳定，返回的数据就是json数据，比较简洁，没有多余的无用内容，方便解析。但是每天的访问次数有一定限制。不过个人使用也够用了。

综上所述，我们选择第二种方式来实现。

所以第一个需要使用的库就是golang官方的net/http库了。

## 有道智云API

网易提供了现有的api，这个api需要先注册，然后获取一个应用的key，同时会生成一个应用的密钥，此处我把这两个东西用appKey和appSecret来表示。至于怎么申请，官方流程会说的很详细

![有道智云应用ID和应用密钥](E:\code\code_go\Godict\doc\img\image_1.png)

## API使用方式

api的使用方式需要参考有道智云的官方文档

[有道智云官方文档]( [https://ai.youdao.com/DOCSIRMA/html/%E8%87%AA%E7%84%B6%E8%AF%AD%E8%A8%80%E7%BF%BB%E8%AF%91/API%E6%96%87%E6%A1%A3/%E6%96%87%E6%9C%AC%E7%BF%BB%E8%AF%91%E6%9C%8D%E5%8A%A1/%E6%96%87%E6%9C%AC%E7%BF%BB%E8%AF%91%E6%9C%8D%E5%8A%A1-API%E6%96%87%E6%A1%A3.html](https://ai.youdao.com/DOCSIRMA/html/自然语言翻译/API文档/文本翻译服务/文本翻译服务-API文档.html) )

从文档上我们获取到一下信息：

文本翻译接口地址: https://openapi.youdao.com/api 

协议：

| 规则     | 描述               |
| -------- | ------------------ |
| 传输方式 | HTTPS              |
| 请求方式 | GET/POST           |
| 字符编码 | 统一使用UTF-8 编码 |
| 请求格式 | 表单               |
| 响应格式 | JSON               |

表单中的参数：

| 字段名   | 类型 | 含义                      | 必填  | 备注                                                         |
| -------- | ---- | ------------------------- | ----- | ------------------------------------------------------------ |
| q        | text | 待翻译文本                | True  | 必须是UTF-8编码                                              |
| from     | text | 源语言                    | True  | 参考下方 [支持语言](https://ai.youdao.com/DOCSIRMA/html/自然语言翻译/API文档/文本翻译服务/文本翻译服务-API文档.html#section-9) (可设置为auto) |
| to       | text | 目标语言                  | True  | 参考下方 [支持语言](https://ai.youdao.com/DOCSIRMA/html/自然语言翻译/API文档/文本翻译服务/文本翻译服务-API文档.html#section-9) (可设置为auto) |
| appKey   | text | 应用ID                    | True  | 可在 [应用管理](https://ai.youdao.com/appmgr.s) 查看         |
| salt     | text | UUID                      | True  | UUID                                                         |
| sign     | text | 签名                      | True  | sha256(应用ID+input+salt+curtime+应用密钥)                   |
| signType | text | 签名类型                  | True  | v3                                                           |
| curtime  | text | 当前UTC时间戳(秒)         | true  | TimeStamp                                                    |
| ext      | text | 翻译结果音频格式，支持mp3 | false | mp3                                                          |
| voice    | text | 翻译结果发音选择          | false | 0为女声，1为男声。默认为女声                                 |

> 签名生成方法如下：
> signType=v3；
> sign=sha256(`应用ID`+`input`+`salt`+`curtime`+`应用密钥`)；
> 其中，input的计算方式为：`input`=`q前10个字符` + `q长度` + `q后10个字符`（当q长度大于20）或 `input`=`q字符串`（当q长度小于等于20）； 

好了，到了这一步基本的一些操作信息就都有了

我们来逐个分析一下参数：

### q

就是需要翻译的文本

### from to

就是从什么语言翻译成什么语言，对应语言的格式官方文档有详细列表。

### appKey

这个就是上面提到的在有道智云申请的弄个东西了。

### salt

是一个uuid，所以我们需要一个用来生成uuid的库，golang有第三方uuid库。

### sign

这个就是最关键的东西，前面，这个需要根据前面的这些信息来计算出来，计算公式上面提到了。

### signType

这个固定位v3就行

## 准备数据

通过上面的分析，可以知道我们需要准备一下数据：

待翻译的单词(word)，uuid，源语言(fromLan)，目标语言(toLan)，appKey，appSecret，sign，signType

我们先考虑一下整个app的工作流程：

鉴于appKey和appSecret是比较私密的东西，所以应该放到配置文件中来让用户配置自己对应的appKey和appSecret，而不应该把这两部分在程序中写死。所以我们需要加载一个配置文件，暂定为应用同级目录下的config.json。

带翻译的单词、源语言和目标语言这个应该是由用户来输入的，所以需要有一个命令行传参，我们借用cobra。

uuid需要在应用程序内实时生成

sign需要根据已知变量来计算，signType固定值

---------------------------------------------------------------------------

## cobra

cobra是一个构建命令行工具的库，我们先大致描述一下我们需要的命令结构，首先word是必须的，还要附加两个标志(flag)：from和to。

所以大概就是这个样子：

```bash
$ ./appname word --from en --to zh-CSH
```

或者简写成

```bash
$ ./appname word -f en -t zh-CSH
```

cobra中的命令组织方式是一个树状的方式，首先有一个根命令，根命令中添加若干个子命令，然后每个子命令又可以添加自己的子命令。

所处cobra中，最基本的单元就是命令(cobra.Command)，命令之间可以添加父子关系，最后组织成一个命令树。

每个命令有基本的5个成员：

- Use 用来描述命令的使用方式
- Short 命令的简短帮助信息
- Long 命令完整的帮助信息
- Args 命令的参数约束
- Run 命令匹配成功后执行的函数体

很显然，我们这个命令工具暂时用不到子命令，所以我们直接使用一个根命令即可。

```go
var rootCmd = &cobra.Command{Use: "app {word}", Short: "translate words",
		Long: `translate words to other language by cmdline`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			//do something
             fmt.Println("Arg:", strings.Join(args, " "))
			fmt.Println("translate:", "from", fromLan, "to", toLan)
		}}
```

根命令按照上面的定义即可。

还有两个flag需要添加到rootCmd上面，这两个选项是以kv键值对形式存在的，可以省略，所以需要提供一个默认值。根据有道智云的文档可以看到，翻译语言可以自动识别，所以我们只需要默认设置成auto即可

```go
rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
```

此处我们需要定义两个字符串变量来接收这两个flag的值

```go
var fromLan string
var toLan string
```

然后只需要将这个命令运行起来即可

```go
rootCmd.Execute()
```

完整代码:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var fromLan string
var toLan string

func main() {
	var rootCmd = &cobra.Command{Use: "app {word}", Short: "translate words",
		Long: `translate words to other language by cmdline`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Arg:", strings.Join(args, " "))
			fmt.Println("translate:", "from", fromLan, "to", toLan)
		}}

	rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
	rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
	rootCmd.Execute()
}
```

```bash
PS E:\code\code_go\Godict\test> go build
PS E:\code\code_go\Godict\test> ./test nice --from en --to zh-CSH
Arg: nice
translate: from en to zh-CSH
```



-----------------------------------------------------

## 加载config.json配置

由于golang自带有json编解码库，所以我们使用json格式的配置文件。

如前面所述，配置文件需要加载appKey和appSecret两个参数，因此定义如下：

```json
{
    "appKey":"your app key",
    "appSecret":"your app secret code"
}
```

json在golang中，使用tag来指定json与结构体的映射

```go
type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}
```

json是一个文本文件，所以我们首先需要把文件中的内容读取出来

```go
fileobj, err := os.Open(str)
if err != nil {
	return err
}

defer fileobj.Close()

var fileContext []byte
fileContext, err = ioutil.ReadAll(fileobj)
```

然后将读取出来的内容使用json.Unmarshal函数解析

```go
json.Unmarshal(fileContext, cfg)
```

我们将此部分代码定义成一个函数方便调用:

```go
func InitConfig(str string, cfg *Config) error {
	fileobj, err := os.Open(str)
	if err != nil {
		return err
	}

	defer fileobj.Close()

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	return nil
}
```

此处函数需要传入config.json文件的路径和解析成功后保存数据的Config变量指针。我们此处规定加载应用同级目录下的config.json。所以我们需要能获取应用程序的绝对路径，此处使用绝对路径是为了保证config.json一定能获取到。

绝对路径可以使用一下方式获取：

```go
dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
```

整理成一个函数方便调用：

```go
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}
```



上述完整代码和测试:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

var config Config
var fromLan string
var toLan string

func main() {
	var rootCmd = &cobra.Command{Use: "app {word}", Short: "translate words",
		Long: `translate words to other language by cmdline`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Arg:", strings.Join(args, " "))
			fmt.Println("translate:", "from", fromLan, "to", toLan)

			var curpath string = GetCurrentDirectory()

			err := InitConfig(curpath+"/config.json", &config)
			if err != nil {
				fmt.Println("config.json is open error.")
				return
			}
			fmt.Println("appKey:", config.AppKey)
			fmt.Println("appSecret:", config.AppSecret)
		}}

	rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
	rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
	rootCmd.Execute()
}

func InitConfig(str string, cfg *Config) error {
	fileobj, err := os.Open(str)
	if err != nil {
		return err
	}

	defer fileobj.Close()

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	return nil
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

```

```bash
PS E:\code\code_go\Godict\test> go build
PS E:\code\code_go\Godict\test> ./test nice --from en --to zh-CSH
Arg: nice
translate: from en to zh-CSH
appKey: your app key
appSecret: your app secret code
```



-------------------------

## 生成UUID

golang有现成的第三方uuid库，使用比较简单

```go
import uuid "github.com/satori/go.uuid"
```

生成uuid

```go
u1 := uuid.NewV4()
u1str := u1.String()
```

由于比较简单，此处就不放完整的测试代码了。

```bash
PS E:\code\code_go\Godict\test> go build
PS E:\code\code_go\Godict\test> ./test nice --from en --to zh-CSH
Arg: nice
translate: from en to zh-CSH
appKey: your app key
appSecret: your app secret code
uuid: 3ad0c54a-24a6-476b-9b8c-730656b5b759
```



-----------

## 计算sign

通过上面的流程，我们已经可以获取到一个查询api需要的大部分参数了，除了sign。所以这一步就是来计算sign，sign就是将前面获取到的值，通过一定规律组合，然后使用指定算法计算出来。

>  签名生成方法如下：
> signType=v3；
> sign=sha256(`应用ID`+`input`+`salt`+`curtime`+`应用密钥`)；
> 其中，input的计算方式为：`input`=`q前10个字符` + `q长度` + `q后10个字符`（当q长度大于20）或 `input`=`q字符串`（当q长度小于等于20）； 

可以看到sign的计算中需要：应用ID(appKey)，input，salt(uuid)，curtime，应用密钥(appSecret)。

这些变量中，只有curtime这个需要还没有获取。

curtime就是当前时间的秒时间戳，利用golang的time库很容易获取：

```go
stamp := time.Now().Unix()
```

但是这个地方获取的是一个int64的整型数值，我们需要转换为字符换。可以利用strconv.FormatInt来转换成字符串。为什么不用os.Itoa？因为os.Itoa的的入参类型为int，而strconv.FormatInt的入参类型为int64，为了确保变量精度一直，所以直接用strconv.FormatInt。

```go
strconv.FormatInt(stamp, 10)
```



可以注意到对于input的处理，input可以说就是精简版的q，精简规则上面有说明：

input的计算方式为：`input`=`q前10个字符` + `q长度` + `q后10个字符`（当q长度大于20）或 `input`=`q字符串`（当q长度小于等于20）；

我们把这个地方提炼成一个函数方便调用：

```go
func truncate(q string) string {
	res := make([]byte, 10)
	qlen := len([]rune(q))
	if qlen <= 20 {
		return q
	} else {
		temp := []byte(q)
		copy(res, temp[:10])
		lenstr := strconv.Itoa(qlen)
		res = append(res, lenstr...)
		res = append(res, temp[qlen-10:qlen]...)
		return string(res)
	}
}
```

至此，所有用来计算sign的变量都准备好了，我们只需要将这些资源的字符串形式拼接起来，然后使用sha256计算即可。sha256的计算直接调用自带的库中的sig := sha256.Sum256。

```go
u1 := uuid.NewV4()
input := truncate(words)
stamp := time.Now().Unix()
instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
sig := sha256.Sum256([]byte(instr))
```

我们成功计算出来sign，但是这是计算出来的结果还是16进制的，而我们实际需求的是字符串格式的，即需要将hex转换成对应的16进制字符串，比如将{0x11,0x56,0xA3}转换成“1156A3”这样的。

我们上面有提及到strconv.FormatInt可以将数字转换成字符串，而且可以指定转换的进制。那么我们只需要将sig这个16进制切片的每个元素转换成对应的字符串，然后拼接起来即可。但是需要注意的是，像0x05这样的用strconv.FormatInt转换出来的字符串会只有一个字符长度，毕竟0x05实际就是0x5。这就不太符合我们的需求了，但是问题不大，人为的判断处理一下即可。此处我们仍然将这个转换写成一个函数：

```go
func HexBuffToString(buff []byte) string {
	var ret string
	for _, value := range buff {
		str := strconv.FormatUint(uint64(value), 16)
		if len([]rune(str)) == 1 {
			ret = ret + "0" + str
		} else {
			ret = ret + str
		}
	}
	return ret
}
```

为了方便测试，我们对appKey和appSecret的值做一下设定:

```json
{
    "appKey":"appKey",
    "appSecret":"appSecret"
}
```

完整的测试代码：

```go
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

var config Config
var fromLan string
var toLan string

func main() {
	var rootCmd = &cobra.Command{Use: "app {word}", Short: "translate words",
		Long: `translate words to other language by cmdline`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			words := strings.Join(args, " ")
			fmt.Println("Arg:", words)
			fmt.Println("translate:", "from", fromLan, "to", toLan)

			var curpath string = GetCurrentDirectory()

			err := InitConfig(curpath+"/config.json", &config)
			if err != nil {
				fmt.Println("config.json is open error.")
				return
			}
			fmt.Println("appKey:", config.AppKey)
			fmt.Println("appSecret:", config.AppSecret)

			u1 := uuid.NewV4()
			fmt.Println("uuid:", u1.String())

			input := truncate(words)
			stamp := time.Now().Unix()
			instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
			sig := sha256.Sum256([]byte(instr))
			var sigstr string = HexBuffToString(sig[:])
			fmt.Println("sign:", sigstr)
		}}

	rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
	rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
	rootCmd.Execute()
}

func InitConfig(str string, cfg *Config) error {
	fileobj, err := os.Open(str)
	if err != nil {
		return err
	}

	defer fileobj.Close()

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	return nil
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func truncate(q string) string {
	res := make([]byte, 10)
	qlen := len([]rune(q))
	if qlen <= 20 {
		return q
	} else {
		temp := []byte(q)
		copy(res, temp[:10])
		lenstr := strconv.Itoa(qlen)
		res = append(res, lenstr...)
		res = append(res, temp[qlen-10:qlen]...)
		return string(res)
	}
}

func HexBuffToString(buff []byte) string {
	var ret string
	for _, value := range buff {
		str := strconv.FormatUint(uint64(value), 16)
		if len([]rune(str)) == 1 {
			ret = ret + "0" + str
		} else {
			ret = ret + str
		}
	}
	return ret
}
```

测试：

```bash
PS E:\code\code_go\Godict\test> go build
PS E:\code\code_go\Godict\test> ./test nice --from en --to zh-CSH
Arg: nice
translate: from en to zh-CSH
appKey: appKey
appSecret: appSecret
uuid: f938ba34-3b59-427d-ac36-779d935a0896
sign: 2f94f435042839a6c5fcb82578c31ff6390f7efeee160365e1be420b23585ee3
```



--------------

## 有道API的POST请求

有道api支持get和post，我们此处使用post方式，post的所有参数前面都已经能够获取了。

golang发起post请求需要引入net/http库

```go
"net/http"
```

因为带有参数，所以是以表单的形式发起请求的，直接使用http.PostForm。该函数需要传入一个url.Values，实际就是一个map。

我们使用前面准备好的数据来构造这个map：

```go
data := make(url.Values, 0)
data["q"] = []string{words}
data["from"] = []string{from}
data["to"] = []string{to}
data["appKey"] = []string{config.AppKey}
data["salt"] = []string{u1.String()}
data["sign"] = []string{sigstr}
data["signType"] = []string{signType}
data["curtime"] = []string{strconv.FormatInt(stamp, 10)}
```

使用http.PostForm发起请求，响应的结果会保存在http.Response中

```go
var resp *http.Response
resp, err = http.PostForm("https://openapi.youdao.com/api",data)
if err != nil {
	fmt.Println(err)
}

defer resp.Body.Close()
```

提取body中的json数据

```go
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
	// handle error
}
```

注意：此时的appKey和appSecret必须要是有效的值，否则无法得到想要的结果

```bash
PS E:\code\code_go\Godict\test> .\test.exe nice -f en -t zh-CSH
Arg: nice
translate: from en to zh-CSH
uuid: f0960478-26f4-47a1-a3d6-fe8f5b28afa6
resp body: {"tSpeakUrl":"http://openapi.youdao.com/ttsapi?q=%E4%B8%8D%E9%94%99%E7%9A%84&langType=zh-CHS&sign=3915602C29F3B9B9B2A521ABABB13D20&salt=1576996331194&voice=4&format=mp3&appKey=4a582e1425d5810e","returnPhrase":["nice"],"web":[{"value":["尼斯","研究所","美好的","英
国国家卫生与临床优化研究所"],"key":"Nice"},{"value":["好人文化","好好先生","老好人"],"key":"nice guy"},{"value":["奈伊茜","奈思河","以市 
场为导向","从顾客需求的角度着手"],"key":"NICE CLAUP"}],"query":"nice","translation":["不错的"],"errorCode":"0","dict":{"url":"yddict://m.youdao.com/dict?le=eng&q=nice"},"webdict":{"url":"http://m.youdao.com/dict?le=eng&q=nice"},"basic":{"us-phonetic":"naɪs","phonetic":"naɪs","uk-phonetic":"naɪs","wfs":[{"wf":{"name":"比较级","value":"nicer"}},{"wf":{"name":"最高级","value":"nicest"}}],"uk-speech":"http://openapi.youdao.com/ttsapi?q=nice&langType=en&sign=151BAD30E03C856BD7154428FA13C367&salt=1576996331194&voice=5&format=mp3&appKey=xxxxxxxxxxx","explains":["adj. 精密的；美好的；细微的；和蔼的","n. (Nice)人名；(英)尼斯"],"us-speech":"http://openapi.youdao.com/ttsapi?q=nice&langType=en&sign=151BAD30E03C856BD7154428FA13C367&salt=1576996331194&voice=6&format=mp3&appKey=xxxxxxxxxxxx"},"l":"en2zh-CHS","speakUrl":"http://openapi.youdao.com/ttsapi?q=nice&langType=en&sign=151BAD30E03C856BD7154428FA13C367&salt=1576996331194&voice=4&format=mp3&appKey=xxxxxxxxxxxxxx"}
```



-----------

## 解析查询的json结果数据

数据解析我们可以参考实际返回的json结果和有道智云的文档说明。

返回的结果是json格式，包含字段与FROM和TO的值有关，具体说明如下：

| 字段名       | 类型  | 含义             | 备注                                                         |
| ------------ | ----- | ---------------- | ------------------------------------------------------------ |
| errorCode    | text  | 错误返回码       | 一定存在                                                     |
| query        | text  | 源语言           | 查询正确时，一定存在                                         |
| translation  | Array | 翻译结果         | 查询正确时，一定存在                                         |
| basic        | text  | 词义             | 基本词典，查词时才有                                         |
| web          | Array | 词义             | 网络释义，该结果不一定存在                                   |
| l            | text  | 源语言和目标语言 | 一定存在                                                     |
| dict         | text  | 词典deeplink     | 查询语种为支持语言时，存在                                   |
| webdict      | text  | webdeeplink      | 查询语种为支持语言时，存在                                   |
| tSpeakUrl    | text  | 翻译结果发音地址 | 翻译成功一定存在，需要应用绑定语音合成实例才能正常播放 否则返回110错误码 |
| speakUrl     | text  | 源语言发音地址   | 翻译成功一定存在，需要应用绑定语音合成实例才能正常播放 否则返回110错误码 |
| returnPhrase | Array | 单词校验后的结果 | 主要校验字母大小写、单词前含符号、中文简繁体                 |

注：

a. 中文查词的basic字段只包含explains字段。

b. 英文查词的basic字段中又包含以下字段。

| 字段        | 含义                                             |
| ----------- | ------------------------------------------------ |
| us-phonetic | 美式音标，英文查词成功，一定存在                 |
| phonetic    | 默认音标，默认是英式音标，英文查词成功，一定存在 |
| uk-phonetic | 英式音标，英文查词成功，一定存在                 |
| uk-speech   | 英式发音，英文查词成功，一定存在                 |
| us-speech   | 美式发音，英文查词成功，一定存在                 |
| explains    | 基本释义                                         |

通过对比，我们发现实际的结果和文档上大体一致，但是dict和webdict这两个字段略有差异，这两个实际返回的是一个对象，而不是文档上的text。

我们现在来定义这个json对应的数据结构

```go
type DictResp struct {
	ErrorCode    string                 `json:"errorCode"`
	Query        string                 `json:"query"`
	Translation  []string               `json:"translation"`
	Basic        DictBasic              `json:"basic"`
	Web          []DictWeb              `json:"web,omitempty"`
	Lang         string                 `json:"l"`
	Dict         map[string]interface{} `json:"dict,omitempty"`
	Webdict      map[string]interface{} `json:"webdict,omitempty"`
	TSpeakUrl    string                 `json:"tSpeakUrl,omitempty"`
	SpeakUrl     string                 `json:"speakUrl,omitempty"`
	ReturnPhrase []string               `json:"returnPhrase,omitempty"`
}
```

ErrorCode、Query、Lang、TSpeakUrl和SpeakUrl是字符串。

Translation和ReturnPhrase是一个字符串数组。

Basic是一个对象，我们直接在里面嵌套一个对应的结构体就行了。

Web是一个对象数组，所以要嵌套一个结构体数组。

Dict和Webdict也是对象，但是它们内部的东西对我们没什么用，我们不需要关心，所以直接定义成map[string]interface{}即可

DictBasic的结构如下：

```go
type DictBasic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}
```

这里面包含着音标和一些拓展的翻译。

DictWeb的数据结构如下：

```go
type DictWeb struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}
```

这里面主要包含的是网络翻译。

之后只需要调用json库的解析函数就行了

```go
var jsonObj DictResp
json.Unmarshal(body, &jsonObj)
```

----------------

## 显示

正确获取了我们想要的结果后，我们只需要按照我们希望的格式显示出来即可，下面提供一个格式化显示函数，以供参考：

```go
func show(resp *DictResp, w io.Writer) {
	if resp.ErrorCode != "0" {
		fmt.Fprintln(w, "请输入正确的数据")
	}
	fmt.Fprintln(w, "@", resp.Query)

	if resp.Basic.UkPhonetic != "" {
		fmt.Fprintln(w, "英:", "[", resp.Basic.UkPhonetic, "]")
	}
	if resp.Basic.UsPhonetic != "" {
		fmt.Fprintln(w, "美:", "[", resp.Basic.UsPhonetic, "]")
	}

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.Translation {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}
	fmt.Fprintln(w, "[延伸]")
	for key, item := range resp.Basic.Explains {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}

	fmt.Fprintln(w, "[网络]")
	for key, item := range resp.Web {
		fmt.Fprintln(w, "\t", key+1, ".", item.Key)
		fmt.Fprint(w, "\t翻译:")
		for _, val := range item.Value {
			fmt.Fprint(w, val, ",")
		}
		fmt.Fprint(w, "\n")
	}
}
```

如果像直接在控制台显示，可以这样调用：

```go
show(&jsonObj, os.Stdout)
```

那么最后完整的代码就是这样的：

```go
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

type DictWeb struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

type DictBasic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}

type DictResp struct {
	ErrorCode    string                 `json:"errorCode"`
	Query        string                 `json:"query"`
	Translation  []string               `json:"translation"`
	Basic        DictBasic              `json:"basic"`
	Web          []DictWeb              `json:"web,omitempty"`
	Lang         string                 `json:"l"`
	Dict         map[string]interface{} `json:"dict,omitempty"`
	Webdict      map[string]interface{} `json:"webdict,omitempty"`
	TSpeakUrl    string                 `json:"tSpeakUrl,omitempty"`
	SpeakUrl     string                 `json:"speakUrl,omitempty"`
	ReturnPhrase []string               `json:"returnPhrase,omitempty"`
}

var config Config
var fromLan string
var toLan string

func main() {
	var rootCmd = &cobra.Command{Use: "app {word}", Short: "translate words",
		Long: `translate words to other language by cmdline`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			words := strings.Join(args, " ")
			var curpath string = GetCurrentDirectory()

			err := InitConfig(curpath+"/config.json", &config)
			if err != nil {
				fmt.Println("config.json is open error.")
				return
			}

			u1 := uuid.NewV4()

			input := truncate(words)
			stamp := time.Now().Unix()
			instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
			sig := sha256.Sum256([]byte(instr))
			var sigstr string = HexBuffToString(sig[:])

			data := make(url.Values, 0)
			data["q"] = []string{words}
			data["from"] = []string{fromLan}
			data["to"] = []string{toLan}
			data["appKey"] = []string{config.AppKey}
			data["salt"] = []string{u1.String()}
			data["sign"] = []string{sigstr}
			data["signType"] = []string{"v3"}
			data["curtime"] = []string{strconv.FormatInt(stamp, 10)}

			var resp *http.Response
			resp, err = http.PostForm("https://openapi.youdao.com/api",
				data)
			if err != nil {
				fmt.Println(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// handle error
			}

			var jsonObj DictResp
			json.Unmarshal(body, &jsonObj)

			show(&jsonObj, os.Stdout)
		}}

	rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
	rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
	rootCmd.Execute()
}

func InitConfig(str string, cfg *Config) error {
	fileobj, err := os.Open(str)
	if err != nil {
		return err
	}

	defer fileobj.Close()

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	return nil
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func truncate(q string) string {
	res := make([]byte, 10)
	qlen := len([]rune(q))
	if qlen <= 20 {
		return q
	} else {
		temp := []byte(q)
		copy(res, temp[:10])
		lenstr := strconv.Itoa(qlen)
		res = append(res, lenstr...)
		res = append(res, temp[qlen-10:qlen]...)
		return string(res)
	}
}

func HexBuffToString(buff []byte) string {
	var ret string
	for _, value := range buff {
		str := strconv.FormatUint(uint64(value), 16)
		if len([]rune(str)) == 1 {
			ret = ret + "0" + str
		} else {
			ret = ret + str
		}
	}
	return ret
}

func show(resp *DictResp, w io.Writer) {
	if resp.ErrorCode != "0" {
		fmt.Fprintln(w, "请输入正确的数据")
	}
	fmt.Fprintln(w, "@", resp.Query)

	if resp.Basic.UkPhonetic != "" {
		fmt.Fprintln(w, "英:", "[", resp.Basic.UkPhonetic, "]")
	}
	if resp.Basic.UsPhonetic != "" {
		fmt.Fprintln(w, "美:", "[", resp.Basic.UsPhonetic, "]")
	}

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.Translation {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}
	fmt.Fprintln(w, "[延伸]")
	for key, item := range resp.Basic.Explains {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}

	fmt.Fprintln(w, "[网络]")
	for key, item := range resp.Web {
		fmt.Fprintln(w, "\t", key+1, ".", item.Key)
		fmt.Fprint(w, "\t翻译:")
		for _, val := range item.Value {
			fmt.Fprint(w, val, ",")
		}
		fmt.Fprint(w, "\n")
	}
}
```

效果：

```bash
PS E:\code\code_go\Godict\test> go build
PS E:\code\code_go\Godict\test> .\test.exe 手机 -f zh-CSH -t en
@ 手机
[翻译]
         1 . Mobile phone
[延伸]
         1 . mobile phone
         2 . cellphone
[网络]
         1 . 手机
        翻译:mobile phone,Iphone,handset,
         2 . 手机电视
        翻译:CMMB,DVB-H,mobile tv,Dopool,
         3 . 翻盖手机
        翻译:flip,clamshell phone,OPPO,flip cell phone,
PS E:\code\code_go\Godict\test>
```

当然-f 和-t也可以省略：

```bash
PS E:\code\code_go\Godict\test> .\test.exe work
@ work
英: [ wɜːk ]
美: [ wɜːrk ]
[翻译]
         1 . 工作
[延伸]
         1 . n. 工作；功；产品；操作；职业；行为；事业；工厂；著作；文学、音乐或艺术作品
         2 . vt. 使工作；操作；经营；使缓慢前进
         3 . vi. 工作；运作；起作用
         4 . n. （Work）（英、埃塞、丹、冰、美）沃克（人名）
[网络]
         1 . Work
        翻译:作品,起作用,工件,运转,
         2 . at work
        翻译:在工作,忙于,上班,在上班,
         3 . work function
        翻译:功函数,逸出功,自由能,功函數,
PS E:\code\code_go\Godict\test>
```

