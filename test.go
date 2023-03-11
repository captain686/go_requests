package gorequests

import "fmt"

func main() {
	//headers := HeaderMap("{ \"User-Agent\": \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36\"}")
	//
	//fmt.Println(headers)
	var req Requests
	req = Req{
		Method:   "GET",
		Host:     "https://httpbin.org/get",
		ProxyUrl: "http://127.0.0.1:8080",
		//Header:   headers,
	}
	response, err := req.Requests()
	if err != nil {
		return
	}

	fmt.Println(*response.Text)
}
