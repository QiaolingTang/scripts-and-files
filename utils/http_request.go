package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type HTTPClient struct {
	username        string //Username for HTTP basic auth.
	password        string //Password for HTTP basic auth
	address         string //Server address.
	path            string
	orgID           string //adds X-Scope-OrgID to API requests for representing tenant ID. Useful for requesting tenant data when bypassing an auth gateway.
	bearerToken     string //adds the Authorization header to API requests for authentication purposes.
	bearerTokenFile string //adds the Authorization header to API requests for authentication purposes.
	retries         int    //How many times to retry each query when getting an error response from Loki.
	queryTags       string //adds X-Query-Tags header to API requests.
	quiet           bool   //Suppress query metadata.
	body            string
	query           string
	method          string
	statusCode      int
}

// retry sets how many times to retry each query
func (c *HTTPClient) retry(retry int) *HTTPClient {
	nc := *c
	nc.retries = retry
	return &nc
}

// withToken sets the token used to do query
func (c *HTTPClient) withToken(bearerToken string) *HTTPClient {
	nc := *c
	nc.bearerToken = bearerToken
	return &nc
}

func (c *HTTPClient) withBasicAuth(username string, password string) *HTTPClient {
	nc := *c
	nc.username = username
	nc.password = password
	return &nc
}

func (c *HTTPClient) withTokenFile(bearerTokenFile string) *HTTPClient {
	nc := *c
	nc.bearerTokenFile = bearerTokenFile
	return &nc
}

func getProxyFromEnv() string {
	var proxy string
	if os.Getenv("http_proxy") != "" {
		proxy = os.Getenv("http_proxy")
	} else if os.Getenv("http_proxy") != "" {
		proxy = os.Getenv("https_proxy")
	}
	return proxy
}

func (c *HTTPClient) do() {
	h, err := c.getRequestHeader()
	if err != nil {
		log.Fatalf("got error when getting http request header: %v", err)
	}
	params := url.Values{}
	if len(c.query) > 0 {
		params.Add("query", c.query)
	}
	res, err := doHTTPRequest(h, c.address, c.path, params.Encode(), c.method, c.quiet, c.retries, bytes.NewReader([]byte(c.body)), c.statusCode)
	if err != nil {
		log.Fatalf("got error when running http request: %v", err)
	}
	if !c.quiet {
		fmt.Printf("the result is:\n%s", string(res))
	}
	fmt.Printf(string(res))
}

func doHTTPRequest(header http.Header, address, path, query, method string, quiet bool, attempts int, requestBody io.Reader, expectedStatusCode int) ([]byte, error) {
	us, err := buildURL(address, path, query)
	if err != nil {
		return nil, err
	}
	if !quiet {
		log.Printf("the request URL is: %s", us)
	}

	req, err := http.NewRequest(strings.ToUpper(method), us, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header = header

	var tr *http.Transport
	proxy := getProxyFromEnv()
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyURL),
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{Transport: tr}

	var resp *http.Response
	success := false

	for attempts > 0 {
		attempts--

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("error sending request: %v", err)
			continue
		}
		if resp.StatusCode != expectedStatusCode {
			buf, _ := io.ReadAll(resp.Body) // nolint
			log.Printf("Error response from server: %s (%v) attempts remaining: %d", string(buf), err, attempts)
			if err := resp.Body.Close(); err != nil {
				log.Print("error closing body", err)
			}
			continue
		}
		success = true
		break
	}
	if !success {
		return nil, fmt.Errorf("run out of attempts while querying the server")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Print("error closing body", err)
		}
	}()
	return io.ReadAll(resp.Body)
}

// buildURL concats a url `http://foo/bar` with a path `/buzz`.
func buildURL(u, p, q string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, p)
	url.RawQuery = q
	return url.String(), nil
}

func (c HTTPClient) getRequestHeader() (http.Header, error) {
	h := make(http.Header)
	if c.username != "" && c.password != "" {
		h.Set(
			"Authorization",
			"Basic "+base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password)),
		)
	}
	h.Set("User-Agent", "http-client")

	if c.orgID != "" {
		h.Set("X-Scope-OrgID", c.orgID)
	}

	if c.queryTags != "" {
		h.Set("X-Query-Tags", c.queryTags)
	}

	if (c.username != "" || c.password != "") && (len(c.bearerToken) > 0 || len(c.bearerTokenFile) > 0) {
		return nil, fmt.Errorf("at most one of HTTP basic auth (username/password), bearer-token & bearer-token-file is allowed to be configured")
	}

	if len(c.bearerToken) > 0 && len(c.bearerTokenFile) > 0 {
		return nil, fmt.Errorf("at most one of the options bearer-token & bearer-token-file is allowed to be configured")
	}

	if c.bearerToken != "" {
		h.Set("Authorization", "Bearer "+c.bearerToken)
	}

	if c.bearerTokenFile != "" {
		b, err := os.ReadFile(c.bearerTokenFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read authorization credentials file %s: %s", c.bearerTokenFile, err)
		}
		bearerToken := strings.TrimSpace(string(b))
		h.Set("Authorization", "Bearer "+bearerToken)
	}
	return h, nil
}

func main() {
	address := flag.String("address", "https://localhost:9200", "the request URL")
	path := flag.String("path", "", "the request path")
	method := flag.String("method", "GET", "the request method")
	query := flag.String("query", "", "the query string")
	body := flag.String("body", "", "the request body")
	expectedStatusCode := flag.Int("code", 200, "the expected status code of respose")
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	token := flag.String("token", "", "token")
	tokenFile := flag.String("token-file", "", "token file path")
	retry := flag.Int("retry", 5, "how many times will the request do when error happens")
	quiet := flag.Bool("quiet", true, "to print the url or not")
	flag.Parse()

	var c HTTPClient
	c.retries = *retry
	c.address = *address
	c.path = *path
	c.query = *query
	c.body = *body
	c.method = *method
	c.statusCode = *expectedStatusCode
	c.quiet = *quiet
	if len(*username) > 0 && len(*password) > 0 {
		c = *c.withBasicAuth(*username, *password)
	}

	if len(*token) > 0 {
		c = *c.withToken(*token)
	}

	if len(*tokenFile) > 0 {
		c = *c.withTokenFile(*tokenFile)
	}
	c.do()

}
