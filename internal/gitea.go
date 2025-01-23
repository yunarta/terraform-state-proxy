package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
)

type GiteaHandler struct {
	server string
}

func NewGiteaHandler(server string) *GiteaHandler {
	return &GiteaHandler{
		server: server,
	}
}

func (h *GiteaHandler) fetchBranchInfo(branch string, request *Request, authHeader string) *GiteaBranchInfo {
	branchURL := constructURL(
		"%s/api/v1/repos/%s/%s/contents/%s?ref=%s",
		h.server, request.Project, request.Repository, request.Path, branch)

	response, err := makeHTTPCall(http.MethodGet, branchURL,
		map[string]string{
			"Authorization": authHeader,
		}, nil)
	if err != nil || response.StatusCode != http.StatusOK {
		log.Println("Failed to fetch branch info")
		return nil
	}
	defer closeGracefully(response.Body)

	branchInfo, err := NewGiteaBranchInfo(response.Body)
	if err != nil {
		log.Println("Failed to parse branch info")
		return nil
	}

	return branchInfo
}

func (h *GiteaHandler) Get(c *gin.Context) {
	log.Println("GET request received")
	var (
		err                error
		request            *Request
		response           *http.Response
		branch, authHeader string
		state              State
		decrypted          []byte
	)

	branch, _, authHeader, request, err = parseCommonInput(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	giteaFileUrl := constructURL("%s/%s/%s/raw/branch/%s/%s",
		h.server,
		request.Project, request.Repository, branch, request.Path,
	)

	response, err = makeHTTPCall("GET", giteaFileUrl,
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
			if encryptionKey, set := os.LookupEnv("TF_STATE_ENCRYPTION_KEY"); set && state.AESIV != "" {
				decrypted, err = state.Decrypt(encryptionKey)
				if err == nil {
					c.Data(http.StatusOK, response.Header.Get("Context-Type"), decrypted)
				} else {
					c.String(http.StatusBadRequest, err.Error())
				}
			} else {
				c.Data(http.StatusOK, response.Header.Get("Context-Type"), bodyBytes)
			}
		} else {
			c.String(http.StatusBadRequest, err.Error())
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"version": 1})
	}
}

func (h *GiteaHandler) Post(c *gin.Context) {
	log.Println("POST request received")
	var (
		err                         error
		request                     *Request
		response                    *http.Response
		branch, encrypt, authHeader string
		state                       *State
	)

	var requestBody []byte
	var parameters map[string]string
	var body []byte
	var bodyBytes []byte

	branch, encrypt, authHeader, request, err = parseCommonInput(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	requestBody, err = c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if encryptionKey, set := os.LookupEnv("TF_STATE_ENCRYPTION_KEY"); encrypt == "yes" && set {
		state, err = NewEncryptedState(encryptionKey, requestBody)
		if err == nil {
			requestBody, err = json.MarshalIndent(state, "", "  ")
		} else {
			c.String(http.StatusBadRequest, err.Error())
		}
	}

	// Create a buffer and a multipart writer
	parameters = map[string]string{
		"content": base64.StdEncoding.EncodeToString([]byte(requestBody)),
		"branch":  branch,
	}

	giteaFileUrl := constructURL("%s/%s/%s/raw/branch/%s/%s",
		h.server,
		request.Project, request.Repository, branch, request.Path,
	)

	response, err = makeHTTPCall("HEAD", giteaFileUrl,
		map[string]string{
			"Authorization": authHeader,
		}, nil)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var method string
	if response.StatusCode == 404 {
		method = "POST"
	} else {
		method = "PUT"
		if branchInfo := h.fetchBranchInfo(branch, request, authHeader); branchInfo != nil {
			if commitId := branchInfo.GetCommitId(); commitId != "" {
				parameters["sha"] = commitId
			}
		}
	}

	body, err = json.Marshal(parameters)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Println(string(body))
	giteaPutUrl := constructURL(
		"%s/api/v1/repos/%s/%s/contents/%s",
		h.server,
		request.Project, request.Repository, request.Path)
	response, err = makeHTTPCall(method, giteaPutUrl,
		map[string]string{
			"Authorization": authHeader,
			"Content-Type":  "application/json; charset=utf-8",
		}, bytes.NewReader(body))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer closeGracefully(response.Body)

	bodyBytes, err = io.ReadAll(response.Body)
	log.Println(fmt.Sprintf("status: >%s<", string(bodyBytes)))

	c.String(response.StatusCode, string(bodyBytes))
}
