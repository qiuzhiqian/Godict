package main

import (
	"crypto/sha256"
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

	"encoding/json"
	"godist/utils"

	uuid "github.com/satori/go.uuid"
)

const (
	rawtext  string = "通用"
	fromlan  string = "zh-CHS"
	tolan    string = "en"
	signType string = "v3"
)

type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

type DistWeb struct {
	Value []string `json:"value"`
	Key   string   `json:"key"`
}

type DistBasic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}

type DistResp struct {
	ErrorCode    string                 `json:"errorCode"`
	Query        string                 `json:"query"`
	Translation  []string               `json:"translation"`
	Basic        DistBasic              `json:"basic"`
	Web          []DistWeb              `json:"web,omitempty"`
	Lang         string                 `json:"l"`
	Dict         map[string]interface{} `json:"dict,omitempty"`
	Webdict      map[string]interface{} `json:"webdict,omitempty"`
	TSpeakUrl    string                 `json:"tSpeakUrl,omitempty"`
	SpeakUrl     string                 `json:"speakUrl,omitempty"`
	ReturnPhrase []string               `json:"returnPhrase,omitempty"`
}

var config Config

func main() {
	var curpath string = GetCurrentDirectory()

	InitConfig(curpath+"/config.json", &config)
	httpPost()
}

func httpPost() {
	var err error
	u1 := uuid.NewV4()
	fmt.Println("u1:", u1)
	input := truncate(rawtext)
	stamp := time.Now().Unix()
	instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
	fmt.Println("input:", input)
	fmt.Println(instr)
	sig := sha256.Sum256([]byte(instr))
	var sigstr string = utils.HexBuffToString(sig[:])
	//for _, item := range sig {
	//	sigstr = sigstr + strconv.FormatInt(int64(item), 16)
	//}
	fmt.Println(sig)
	fmt.Println(sigstr)

	data := make(url.Values, 0)
	data["q"] = []string{rawtext}
	data["from"] = []string{fromlan}
	data["to"] = []string{tolan}
	data["appKey"] = []string{config.AppKey}
	data["salt"] = []string{u1.String()}
	data["sign"] = []string{sigstr}
	data["signType"] = []string{signType}
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

	fmt.Println(string(body))
	var jsonObj DistResp
	json.Unmarshal(body, &jsonObj)
	fmt.Println(jsonObj)

	show(&jsonObj, os.Stdout)
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

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func InitConfig(str string, cfg *Config) {
	fileobj, err := os.Open(str)
	if err != nil {
		fmt.Println("error:", err)
	}

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	fmt.Println(*cfg)
}

func show(resp *DistResp, w io.Writer) {
	fmt.Fprintln(w, resp.Query)

	if resp.Basic.UkPhonetic != "" {
		fmt.Fprintln(w, "英:", "[", resp.Basic.UkPhonetic, "]")
	}
	if resp.Basic.UsPhonetic != "" {
		fmt.Fprintln(w, "美:", "[", resp.Basic.UsPhonetic, "]")
	}

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.Translation {
		fmt.Fprintln(w, "\t", key, ".", item)
	}
	fmt.Fprintln(w, "[延伸]")
	for key, item := range resp.Basic.Explains {
		fmt.Fprintln(w, "\t", key, ".", item)
	}

	fmt.Fprintln(w, "[网络]")
	for key, item := range resp.Web {
		fmt.Fprintln(w, "\t", key, ".", item.Key)
		fmt.Fprint(w, "\t解释:")
		for _, val := range item.Value {
			fmt.Fprint(w, val, ",")
		}
		fmt.Fprint(w, "\n")
	}
}
