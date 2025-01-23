package internal

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

type Closeable interface {
	Close() error
}

func closeGracefully(body Closeable) {
	_ = body.Close()
}

func parseCommonInput(c *gin.Context) (string, string, *Request, error) {
	branch := c.Query("branch")
	encrypt := c.Query("encrypt")
	authHeader := c.GetHeader("Authorization")

	request := Request{}
	err := c.ShouldBindUri(&request)
	if err != nil {
		c.String(http.StatusBadRequest, "")
		return "", "", nil, err
	}

	request.Path = strings.TrimPrefix(request.Path, "/")
	return branch, encrypt, authHeader, &request, nil
}

func constructURL(format string, server string, params ...interface{}) string {
	return fmt.Sprintf(format, append([]interface{}{strings.TrimRight(server, "/")}, params...)...)
}

func makeHTTPCall(method, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	log.Printf("Initiating HTTP request: Method=%s, URL=%s", method, url)

	req, err = http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, err
	}

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error performing HTTP request: %v", err)
		return nil, err
	}

	log.Printf("HTTP request completed: StatusCode=%d", resp.StatusCode)
	return resp, nil
}

func makeMultipart(parameters map[string]string) (*bytes.Buffer, string, error) {
	var (
		writer      *multipart.Writer
		err         error
		body        *bytes.Buffer
		contentType string
	)

	body = new(bytes.Buffer)

	writer = multipart.NewWriter(body)
	defer closeGracefully(writer)

	for key, value := range parameters {
		err = writer.WriteField(key, value)
		if err != nil {
			return nil, "", nil
		}
	}

	contentType = writer.FormDataContentType()
	return body, contentType, err
}
