package cas

import (
	"bytes"
	_ "embed"
	"errors"
	"github.com/dop251/goja"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

var ltRegexp = regexp.MustCompile("<input type=\"hidden\" id=\"lt\" name=\"lt\" value=\"(.*)\" />")
var executionRegexp = regexp.MustCompile("<input type=\"hidden\" name=\"execution\" value=\"(.*)\" />")

//go:embed des.js
var rawJs []byte

func GenLoginReq(URL, user, passwd string) (*http.Request, error) {
	var lt, execution []byte
	//搞到lt和execution
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tmp := ltRegexp.FindSubmatch(body)
	if len(tmp) != 2 {
		return nil, errors.New("提取lt错误")
	}
	lt = tmp[1]
	tmp = executionRegexp.FindSubmatch(body)
	if len(tmp) != 2 {
		return nil, errors.New("提取execution错误")
	}
	execution = tmp[1]
	//获取rsa
	rsa, err := getRsa(user + passwd + string(lt))
	if err != nil {
		return nil, err
	}

	postData := url.Values{}
	postData.Add("rsa", rsa)
	postData.Add("ul", strconv.Itoa(len(user)))
	postData.Add("pl", strconv.Itoa(len(passwd)))
	postData.Add("lt", string(lt))
	postData.Add("execution", string(execution))
	postData.Add("_eventId", "submit")
	req, err = http.NewRequest(http.MethodPost, URL, bytes.NewBufferString(postData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36")
	req.Header.Add("Referer", URL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range resp.Cookies() {
		req.AddCookie(c)
	}
	if err != nil {
		return nil, err
	}
	return req, nil
}
func getRsa(data string) (string, error) {
	vm := goja.New()
	_, err := vm.RunString(string(rawJs))
	if err != nil {
		panic(err)
	}
	strEnc, valid := goja.AssertFunction(vm.Get("strEnc"))
	if !valid {
		return "", errors.New("invalid js")
	}
	value, err := strEnc(nil, vm.ToValue(data), vm.ToValue("1"), vm.ToValue("2"), vm.ToValue("3"))
	if err != nil {
		panic(err)
	}
	var result = value.String()
	return result, nil
}
