package DouYu

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

func (l *Link) getRateStream() (gjson.Result, error) {
	z := html.NewTokenizer(strings.NewReader(l.res))
	var scripts []string
	var stag bool
FIND:
	for {
		t := z.Next()
		switch t {
		case html.ErrorToken:
			break FIND
		case html.StartTagToken:
			x := z.Token()
			stag = x.Data == "script"
		case html.TextToken:
			x := z.Token()
			if stag {
				scripts = append(scripts, x.Data)
			}
			stag = false
		}
	}
	var jj string
	for _, c := range scripts {
		if strings.Contains(c, "ub98484234") {
			jj = c
		}
	}
	if jj == "" {
		return gjson.Result{}, fmt.Errorf("could not find ub98484234 function")
	}

	// Step 2: Modify the JavaScript function to remove eval statement
	replaceRe := regexp.MustCompile(`return\s*eval\(strc\)\([0-9a-z]+,[0-9a-z]+,[0-9a-z]+\);}`)
	jsFunc := replaceRe.ReplaceAllString(jj, "return strc;}")

	// Step 3: Compile and execute the modified JavaScript function
	js := fmt.Sprintf("%s\nub98484234()", jsFunc)

	vm := goja.New()
	vmResult, err := vm.RunString(js)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("running ub98484234 error: %w", err)
	}
	result := vmResult.String()
	re := regexp.MustCompile(`v=(\d+)`)
	match := re.FindStringSubmatch(result)
	if len(match) < 2 {
		return gjson.Result{}, fmt.Errorf("could not find v parameter")
	}
	v := match[1]
	// Step 5: Generate rb parameter using md5 function
	rbByte := md5.Sum([]byte(fmt.Sprintf("%s%s%s%s", l.rid, l.did, l.t10, v)))
	rb := hex.EncodeToString(rbByte[:])

	// Step 6: Modify JavaScript function to replace return statement with rb parameter
	jsFunc = strings.Replace(result, "return rt;})", "return rt;}", -1)
	jsFunc = strings.Replace(jsFunc, "(function (", "function sign(", -1)
	jsFunc = strings.Replace(jsFunc, "CryptoJS.MD5(cb).toString()", fmt.Sprintf("\"%s\"", rb), -1)
	jsFunc = fmt.Sprintf("%s sign(%s, \"%s\", %s);", jsFunc, l.rid, l.did, l.t10)

	vmSignResult, err := vm.RunString(jsFunc)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("running signature function error: %w", err)
	}
	result = vmSignResult.String()
	params := fmt.Sprintf("%s&ver=Douyu_223092005&rate=0&cdn=&iar=0&ive=0&hevc=0&fa=0", result)
	url := fmt.Sprintf("https://playweb.douyu.com/lapi/live/getH5Play/%s", l.rid)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(params))
	if err != nil {
		return gjson.Result{}, fmt.Errorf("making RateStream POST request error: %w", err)
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("sending RateStream POST request error: %w", err)
	}
	defer resp.Body.Close()

	// Step 8: Parse the JSON response and return it
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("getting RateStream response error: %w", err)
	}

	if l.debug {
		log.Printf("RateStream body: \n%s", string(body))
	}
	return gjson.ParseBytes(body), nil
}
