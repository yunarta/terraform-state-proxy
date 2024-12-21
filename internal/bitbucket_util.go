package internal

import (
	"encoding/json"
	"io"
)

type BitbucketBranchInfo struct {
	Size       int  `json:"size"`
	Limit      int  `json:"limit"`
	IsLastPage bool `json:"isLastPage"`
	Values     []struct {
		Id              string `json:"id"`
		DisplayId       string `json:"displayId"`
		Type            string `json:"type"`
		LatestCommit    string `json:"latestCommit"`
		LatestChangeset string `json:"latestChangeset"`
		IsDefault       bool   `json:"isDefault"`
	} `json:"values"`
	Start int `json:"start"`
}

func (b *BitbucketBranchInfo) GetCommitId() string {
	if len(b.Values) > 0 {
		commitId := b.Values[0].LatestCommit
		return commitId
	}

	return ""
}

func NewBitbucketBranchInfo(body io.ReadCloser) (*BitbucketBranchInfo, error) {
	var branchInfo BitbucketBranchInfo
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
