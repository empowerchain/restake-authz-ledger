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
	addKeyCommand := fmt.Sprintf("$ %s keys add ledger --ledger --keyring-backend file\n", n.CliName)

	minGasPrice := "CHANGEFEE"
	denom := "changedenom"
	if len(n.Fees.FeeTokens) != 0 {
		minGasPrice = fmt.Sprintf("%v", n.Fees.FeeTokens[0].FixedMinGasPrice)
		denom = n.Fees.FeeTokens[0].Denom
	}
	command1 := fmt.Sprintf("$ %s tx authz grant %s generic --msg-type /cosmos.staking.v1beta1.MsgDelegate --from ledger --ledger --chain-id %s --node %s --keyring-backend file --gas auto --gas-prices %s%s --gas-adjustment 1.5",
		n.CliName,
		validator.RestakeAddress,
		n.ChainID,
		n.NodeURI,
		minGasPrice,
		denom)
	command2 := fmt.Sprintf("$ %s tx authz grant %s generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward --from ledger --ledger --chain-id %s --node %s --keyring-backend file --gas auto --gas-prices %v%s --gas-adjustment 1.5",
		n.CliName,
		validator.RestakeAddress,
		n.ChainID,
		n.NodeURI,
		minGasPrice,
		denom)

	if n.Codebase.GitRepo != "" {
		fmt.Printf("You can find the binaries, or build from source here: %s (reported recommended version is %s)\n", n.Codebase.GitRepo, n.Codebase.RecommendedVersion)
		fmt.Println()
	}

	fmt.Printf("First, you need to add you ledger to your keys in %s with the following command:\n%s", n.CliName, addKeyCommand)

	fmt.Println("The following commands will grant your validator access to withdraw rewards and delegate them:")
	fmt.Println()
	fmt.Println(command1)
	fmt.Println()
	fmt.Println(command2)
	fmt.Println("NB! This script is not perfect, and the commands listed above might not work without some small adjustments in certain cases.")
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
