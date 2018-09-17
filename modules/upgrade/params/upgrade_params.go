package upgradeparams

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/irisnet/irishub/modules/parameter"
)

var CurrentUpgradeProposalIdParameter CurrentUpgradeProposalIdParam

var _ parameter.SignalParameter = (*CurrentUpgradeProposalIdParam)(nil)

type CurrentUpgradeProposalIdParam struct {
	Value   int64
	psetter params.Setter
	pgetter params.Getter
}

func (param *CurrentUpgradeProposalIdParam) InitGenesis(genesisState interface{}) {
	param.Value = -1
}

func (param *CurrentUpgradeProposalIdParam) SetReadWriter(setter params.Setter) {
	param.psetter = setter
	param.pgetter = setter.Getter
}

func (param *CurrentUpgradeProposalIdParam) GetStoreKey() string {
	return "Sig/upgrade/proposalId"
}

func (param *CurrentUpgradeProposalIdParam) SaveValue(ctx sdk.Context) {
	param.psetter.Set(ctx, param.GetStoreKey(), param.Value)
}

func (param *CurrentUpgradeProposalIdParam) LoadValue(ctx sdk.Context) bool {
	err := param.pgetter.Get(ctx, param.GetStoreKey(), &param.Value)
	if err != nil {
		return false
	}
	return true
}

func (param *CurrentUpgradeProposalIdParam) ToJson() string {
	jsonBytes, _ := json.Marshal(param.Value)
	return string(jsonBytes)
}

func (param *CurrentUpgradeProposalIdParam) Update(ctx sdk.Context, jsonStr string) {
	if err := json.Unmarshal([]byte(jsonStr), &param.Value); err == nil {
		param.SaveValue(ctx)
	}
}

func (param *CurrentUpgradeProposalIdParam) Valid(jsonStr string) sdk.Error {
	var err error
	if err = json.Unmarshal([]byte(jsonStr), &param.Value); err == nil {
		return nil
	}
	return sdk.NewError(parameter.DefaultCodespace, parameter.CodeInvalidCurrentUpgradeProposalID, fmt.Sprintf("Json is not valid"))
}
