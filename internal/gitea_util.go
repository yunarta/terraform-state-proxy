package internal

import (
	"encoding/json"
	"io"
)

type GiteaBranchInfo struct {
	Links struct {
		Git  string `json:"git"`
		Html string `json:"html"`
		Self string `json:"self"`
	} `json:"_links"`
	Content         string `json:"content"`
	DownloadUrl     string `json:"download_url"`
	Encoding        string `json:"encoding"`
	GitUrl          string `json:"git_url"`
	HtmlUrl         string `json:"html_url"`
	LastCommitSha   string `json:"last_commit_sha"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Sha             string `json:"sha"`
	Size            int    `json:"size"`
	SubmoduleGitUrl string `json:"submodule_git_url"`
	Target          string `json:"target"`
	Type            string `json:"type"`
	Url             string `json:"url"`
}

func (b *GiteaBranchInfo) GetCommitId() string {
	return b.Sha
}

func NewGiteaBranchInfo(body io.ReadCloser) (*GiteaBranchInfo, error) {
	var branchInfo GiteaBranchInfo
	var bytes []byte
	var err error

	bytes, err = io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &branchInfo); err != nil {
		return nil, err
	}

	return &branchInfo, nil
}
