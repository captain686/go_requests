package gorequests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Requests interface {
	Requests() (response, error)
}

type Req struct {
	Method   string
	Host     string
	Data     []byte
	ProxyUrl string
	Header   map[string]string
	Redirect bool
}

type response struct {
	statusCode *int
	text       *string
	raw        *[]byte
}

func randomUa() string {
	ua := browser.Random()
	return ua
}

func hostFormat(host string) string {
	re := regexp.MustCompile("http*.://.*?/")
	host = re.FindString(host)
	return host
}

func HeaderMap(jsonData ...string) map[string]string {
	var headers = make(map[string]string)
	headers["User-Agent"] = randomUa()
	headers["Connection"] = "close"
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	if len(jsonData) > 0 {
		for _, arg := range jsonData {
			err := json.Unmarshal([]byte(arg), &headers)
			if err != nil {
				log.Error(err)
				return headers
			}
		}
	}
	//headersMap := headerMap{headers: headers}
	return headers
}

// (*int, *[]byte, error)
func (requests Req) Requests() (response, error) {

	//代理
	//proxyUrl := "http://127.0.0.1:8080"
	var client = &http.Client{Timeout: time.Second * 15}

	if requests.ProxyUrl != "" {
		proxy, _ := url.Parse(requests.ProxyUrl)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 15, //超时时间
		}
	}

	req, err := http.NewRequest(requests.Method, requests.Host, bytes.NewReader(requests.Data))
	if err != nil {
		log.Error(err)
		return response{}, err
	}

	headers := requests.Header
	if len(headers) == 0 {
		headers = HeaderMap()
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if !requests.Redirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("\n[O] disable Redirect")
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		if !strings.Contains(fmt.Sprint(err), "disable Redirect") {
			log.Error(fmt.Sprintf("%s | err", hostFormat(requests.Host)))
		}
		return response{}, err
	}
	if resp.StatusCode != 200 {
		log.Info(fmt.Sprintf("%s | Status_Code: %d", hostFormat(requests.Host), resp.StatusCode))
	}
	respBody, err := io.ReadAll(resp.Body)
	respText := string(respBody)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if err != nil && respBody == nil {
		log.Error(fmt.Sprintf("%s | %s", hostFormat(requests.Host), err))
		return response{statusCode: &resp.StatusCode}, err
	}
	return response{
		statusCode: &resp.StatusCode,
		text:       &respText,
		raw:        &respBody,
	}, nil
}
