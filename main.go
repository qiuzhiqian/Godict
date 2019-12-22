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

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

const (
	signType string = "v3"
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
			var curpath string = GetCurrentDirectory()

			wordContext := strings.Join(args, " ")
			err := InitConfig(curpath+"/config.json", &config)
			if err != nil {
				fmt.Println("config.json is open error.")
				return
			}
			httpPost(wordContext, fromLan, toLan)
		}}

	rootCmd.Flags().StringVarP(&fromLan, "from", "f", "auto", "translate from this language")
	rootCmd.Flags().StringVarP(&toLan, "to", "t", "auto", "translate to this language")
	rootCmd.Execute()
}

func httpPost(words, from, to string) {
	var err error
	u1 := uuid.NewV4()
	input := truncate(words)
	stamp := time.Now().Unix()
	instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
	sig := sha256.Sum256([]byte(instr))
	var sigstr string = HexBuffToString(sig[:])

	data := make(url.Values, 0)
	data["q"] = []string{words}
	data["from"] = []string{from}
	data["to"] = []string{to}
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

	//fmt.Println(string(body))
	var jsonObj DictResp
	json.Unmarshal(body, &jsonObj)
	//fmt.Println(jsonObj)

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
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
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
