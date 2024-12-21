package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
)

type BitbucketHandler struct {
	server string
}

func NewBitbucketHandler(server string) *BitbucketHandler {
	return &BitbucketHandler{
		server: server,
	}
}

func (h *BitbucketHandler) fetchBranchInfo(branch string, request *Request, authHeader string) *BitbucketBranchInfo {
	branchURL := constructURL(
		"%s/rest/api/1.0/projects/%s/repos/%s/branches?filterText=%s",
		h.server, request.Project, request.Repository, branch)

	response, err := makeHTTPCall(http.MethodGet, branchURL,
		map[string]string{
			"Authorization": authHeader,
		}, nil)
	if err != nil || response.StatusCode != http.StatusOK {
		log.Println("Failed to fetch branch info")
		return nil
	}
	defer closeGracefully(response.Body)

	branchInfo, err := NewBitbucketBranchInfo(response.Body)
	if err != nil {
		log.Println("Failed to parse branch info")
		return nil
	}

	return branchInfo
}

func (h *BitbucketHandler) Get(c *gin.Context) {
	log.Println("GET request received")
	var (
		err                error
		request            *Request
		response           *http.Response
		branch, authHeader string
		state              State
		decrypted          []byte
		bodyBytes          []byte
	)

	branch, authHeader, request, err = parseCommonInput(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	bitbucketFileUrl := constructURL("%s/projects/%s/repos/%s/raw/%s?at=%s",
		h.server,
		request.Project, request.Repository, request.Path, branch,
	)

	response, err = makeHTTPCall("GET", bitbucketFileUrl,
		map[string]string{
			"Authorization": authHeader,
		}, nil)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer closeGracefully(response.Body)

	if response.StatusCode == 200 {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		state = State{}
		if err = json.Unmarshal(bodyBytes, &state); err == nil {
			if encryptionKey, set := os.LookupEnv("TF_STATE_ENCRYPTION_KEY"); set {
				decrypted, err = state.Decrypt(encryptionKey)
				if err == nil {
					c.Data(http.StatusOK, response.Header.Get("Context-Type"), decrypted)
				} else {
					c.String(http.StatusBadRequest, err.Error())
				}
			}
		} else {
			c.Data(http.StatusOK, response.Header.Get("Context-Type"), bodyBytes)
		}
	} else if response.StatusCode == 404 {
		c.JSON(http.StatusOK, gin.H{"version": 1})
	} else {
		bodyBytes, err = io.ReadAll(response.Body)
		log.Println(fmt.Sprintf("status: >%s<", string(bodyBytes)))

		c.String(response.StatusCode, string(bodyBytes))
	}
}

func (h *BitbucketHandler) Post(c *gin.Context) {
	log.Println("POST request received")
	var (
		err                error
		request            *Request
		response           *http.Response
		branch, authHeader string
		state              *State
	)

	var requestBody []byte
	var parameters map[string]string
	var body *bytes.Buffer
	var bodyBytes []byte
	var contentType string

	branch, authHeader, request, err = parseCommonInput(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	requestBody, err = c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if encryptionKey, set := os.LookupEnv("TF_STATE_ENCRYPTION_KEY"); set {
		state, err = NewEncryptedState(encryptionKey, requestBody)
		if err == nil {
			requestBody, err = json.MarshalIndent(state, "", "  ")
		} else {
			c.String(http.StatusBadRequest, err.Error())
		}
	}

	// Create a buffer and a multipart writer
	parameters = map[string]string{
		"content":      string(requestBody),
		"branch":       branch,
		"sourceBranch": branch,
	}

	bitbucketPutUrl := constructURL(
		"%s/rest/api/1.0/projects/%s/repos/%s/browse/%s",
		h.server,
		request.Project, request.Repository, request.Path)

	response, err = makeHTTPCall("HEAD", bitbucketPutUrl,
		map[string]string{
			"Authorization": authHeader,
		}, nil)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if response.StatusCode == 200 {
		if branchInfo := h.fetchBranchInfo(branch, request, authHeader); branchInfo != nil {
			if commitId := branchInfo.GetCommitId(); commitId != "" {
				parameters["sourceCommitId"] = commitId
			}
		}
	}

	body, contentType, err = makeMultipart(parameters)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	response, err = makeHTTPCall("PUT", bitbucketPutUrl,
		map[string]string{
			"Authorization": authHeader,
			"Content-Type":  contentType,
		}, body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer closeGracefully(response.Body)

	bodyBytes, err = io.ReadAll(response.Body)
	log.Println(fmt.Sprintf("status: >%s<", string(bodyBytes)))

	c.String(response.StatusCode, string(bodyBytes))
}
