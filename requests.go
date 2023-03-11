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
	NoVerify bool
}

type response struct {
	StatusCode *int
	Text       *string
	Raw        *[]byte
}

func randomUa() string {
	ua := browser.Random()
	return ua
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

// Requests (*int, *[]byte, error)
func (requests Req) Requests() (response, error) {

	//代理
	//proxyUrl := "http://127.0.0.1:8080"
	// var client = &http.Client{Timeout: time.Second * 15}
	tr := &http.Transport{}
	if requests.ProxyUrl != "" {
		proxy, _ := url.Parse(requests.ProxyUrl)

		if !requests.NoVerify {
			tr = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		} else {
			tr = &http.Transport{
				Proxy:           http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	} else {
		if requests.NoVerify {
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 15, //超时时间
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
			log.Error(fmt.Sprintf("118 %s | %v", requests.Host, err))
		}
		return response{}, err
	}
	if resp.StatusCode != 200 {
		log.Info(fmt.Sprintf("123 %s | Status_Code: %d", requests.Host, resp.StatusCode))
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
		log.Error(fmt.Sprintf("134 %s | %v", requests.Host, err))
		return response{StatusCode: &resp.StatusCode}, err
	}
	return response{
		StatusCode: &resp.StatusCode,
		Text:       &respText,
		Raw:        &respBody,
	}, nil
}
