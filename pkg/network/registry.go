package network

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"net/http"
)

type ChainResponseJSON struct {
	ChainName    string       `json:"chain_name"`
	ChainId      string       `json:"chain_id"`
	PrettyName   string       `json:"pretty_name"`
	Bech32Prefix string       `json:"bech32_prefix"`
	DaemonName   string       `json:"daemon_name"`
	Fees         FeesJSON     `json:"fees"`
	Codebase     CodebaseJSON `json:"codebase"`
	Apis         struct {
		Rpc []struct {
			Address  string `json:"address"`
			Provider string `json:"provider"`
		} `json:"rpc"`
		Rest []struct {
			Address  string `json:"address"`
			Provider string `json:"provider"`
		} `json:"rest"`
		Grpc []struct {
			Address  string `json:"address"`
			Provider string `json:"provider"`
		} `json:"grpc"`
	} `json:"apis"`
}

type FeesJSON struct {
	FeeTokens []struct {
		Denom            string      `json:"denom"`
		FixedMinGasPrice interface{} `json:"fixed_min_gas_price"`
	} `json:"fee_tokens"`
}

type CodebaseJSON struct {
	GitRepo            string   `json:"git_repo"`
	RecommendedVersion string   `json:"recommended_version"`
	CompatibleVersions []string `json:"compatible_versions"`
	Binaries           struct {
		LinuxAmd64   string `json:"linux/amd64"`
		LinuxArm64   string `json:"linux/arm64"`
		DarwinAmd64  string `json:"darwin/amd64"`
		WindowsAmd64 string `json:"windows/amd64"`
	} `json:"binaries"`
}

func GetChainInfo(chainName string) (ChainResponseJSON, error) {
	url := "https://cosmos-chain.directory/chains/" + chainName
	res, err := http.Get(url)
	if err != nil {
		return ChainResponseJSON{}, errors.Wrap(err, 0)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return ChainResponseJSON{}, nil
	}

	if res.StatusCode != 200 {
		return ChainResponseJSON{}, errors.Errorf("Getting chain info failed with a %d code (%s), url: %s", res.StatusCode, res.Status, url)
	}

	chainInfo := ChainResponseJSON{}
	err = json.NewDecoder(res.Body).Decode(&chainInfo)
	if err != nil {
		return ChainResponseJSON{}, errors.Wrap(err, 0)
	}

	return chainInfo, nil
}
