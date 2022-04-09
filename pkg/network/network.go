package network

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bjaanes/restake-authz-ledger/pkg/validator"
	"github.com/cosmos/cosmos-sdk/client"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/go-errors/errors"
	"net/http"
	"net/url"
	"strings"
)

type Network struct {
	Name       string
	Identifier string
	NodeURI    string
	ChainID    string
	CliName    string
	Codebase   CodebaseJSON
	Fees       FeesJSON
}

func GetNetworks() ([]Network, error) {
	res, err := http.Get("https://validators.cosmos.directory")
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Errorf("Getting validators failed with a %d code (%s)", res.StatusCode, res.Status)
	}

	validatorsRes := validator.ValidatorsResponseJSON{}
	err = json.NewDecoder(res.Body).Decode(&validatorsRes)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	foundChains := make(map[string]bool)
	var networks []Network
	for _, v := range validatorsRes.Validators {
		for _, c := range v.Chains {
			if !foundChains[c.Name] {
				foundChains[c.Name] = true
				chainInfo, err := GetChainInfo(c.Name)
				if err != nil {
					fmt.Println(c)
					return nil, errors.Wrap(err, 0)
				}

				if chainInfo.ChainName == "" {
					continue
				}

				nodeURI := chainInfo.Apis.Rpc[0].Address
				u, err := url.ParseRequestURI(nodeURI)
				if err != nil {
					return nil, errors.Wrap(err, 0)
				}
				if u.Port() == "" {
					nodeURI = strings.TrimSuffix(nodeURI, "/")
					nodeURI = nodeURI + ":443"
				}

				daemonName := chainInfo.DaemonName
				if daemonName == "" {
					daemonName = chainInfo.ChainName + "d" // Best effort :shrug:
				}

				networks = append(networks, Network{
					Name:       chainInfo.PrettyName,
					Identifier: chainInfo.ChainName,
					NodeURI:    nodeURI,
					ChainID:    chainInfo.ChainId,
					CliName:    daemonName,
					Codebase:   chainInfo.Codebase,
					Fees:       chainInfo.Fees,
				})
			}
		}
	}

	return networks, nil
}

func (n Network) GrantRestake(addr string, validator validator.ValidatorForNetwork) error {
	addKeyCommand := fmt.Sprintf("$ /path/to/binary/%s keys add ledger --ledger --keyring-backend file\n", n.CliName)

	minGasPrice := "CHANGEFEE"
	denom := "changedenom"
	if len(n.Fees.FeeTokens) != 0 {
		minGasPrice = fmt.Sprintf("%v", n.Fees.FeeTokens[0].FixedMinGasPrice)
		denom = n.Fees.FeeTokens[0].Denom
	}
	withdrawCommand := fmt.Sprintf("$ /path/to/binary/%s tx authz grant %s generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward --from ledger --ledger --chain-id %s --node %s --keyring-backend file --gas auto --gas-prices %v%s --gas-adjustment 1.5",
		n.CliName,
		validator.RestakeAddress,
		n.ChainID,
		n.NodeURI,
		minGasPrice,
		denom)
	delegateCommand := fmt.Sprintf("$ /path/to/binary/%s tx authz grant %s generic --msg-type /cosmos.staking.v1beta1.MsgDelegate --from ledger --ledger --chain-id %s --node %s --keyring-backend file --gas auto --gas-prices %s%s --gas-adjustment 1.5",
		n.CliName,
		validator.RestakeAddress,
		n.ChainID,
		n.NodeURI,
		minGasPrice,
		denom)

	fmt.Println("Instructions to create the necessary authz grants to enable REStake:")
	fmt.Println()
	if n.Codebase.GitRepo != "" {
		fmt.Println("First, you need to get the cli for the chain.")
		fmt.Printf("If you want to build the cli from source, the source code is here: %s (reported recommended version is %s)\n", n.Codebase.GitRepo, n.Codebase.RecommendedVersion)

		if n.Codebase.Binaries.LinuxAmd64 != "" {
			fmt.Printf("Linux amd64: %s\n", n.Codebase.Binaries.LinuxAmd64)
			if n.Codebase.Binaries.LinuxArm64 != "" {
				fmt.Printf("Linux arm64: %s\n", n.Codebase.Binaries.LinuxArm64)
			}
			if n.Codebase.Binaries.DarwinAmd64 != "" {
				fmt.Printf("Darwin (macOS) amd64: %s\n", n.Codebase.Binaries.DarwinAmd64)
			}
			if n.Codebase.Binaries.WindowsAmd64 != "" {
				fmt.Printf("Windows amd64: %s\n", n.Codebase.Binaries.WindowsAmd64)
			}
		} else {
			fmt.Printf("No binaries were reported from the chain registry repo, but you can go to the GitHub repo linked above to find them (usually)")
		}

	} else {
		fmt.Println("You need to find and download the cli for the chain, but we we were not able to find them right now from the chain registry.")
	}
	fmt.Println("!IMPORTANT! In some cases, the binaries are not built with support for Ledger. In those cases you will need to build from source.")
	fmt.Println()

	fmt.Println("After getting the cli, you need to run 3 commands:")
	fmt.Println()
	fmt.Printf("1: Add your ledger to your keys in %s with the following command:\n", n.CliName)
	fmt.Println(addKeyCommand)
	fmt.Println()

	fmt.Println("2: Grant your validator access to withdraw rewards (to your own wallet, not theirs) on your behalf:")
	fmt.Println(withdrawCommand)
	fmt.Println()

	fmt.Println("3: Grant your validator access to delegate on your behalf:")
	fmt.Println(delegateCommand)
	fmt.Println()

	fmt.Println("NB! This script is not perfect, and the commands listed above might not work without some small adjustments in certain cases.")
	fmt.Println("You can check out this cli's accompanying blog post here: https://gjermund.tech/blog/making-ledger-work-on-restake/")
	fmt.Println("Reach out to gjermund#1586 on ECO Stake discord if you're having problems")

	return nil
}

func (n Network) GetDelegations(ctx context.Context, delegatorAddr string) ([]stakingtypes.Delegation, error) {
	c, err := client.NewClientFromNode(n.NodeURI)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	clientCtx := client.Context{
		Offline: false,
		NodeURI: n.NodeURI,
		Client:  c,
	}
	stakingClient := stakingtypes.NewQueryClient(clientCtx)
	params := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegatorAddr,
	}
	res, err := stakingClient.DelegatorDelegations(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	var delegations []stakingtypes.Delegation
	for _, r := range res.GetDelegationResponses() {
		delegations = append(delegations, r.Delegation)
	}

	return delegations, nil
}
