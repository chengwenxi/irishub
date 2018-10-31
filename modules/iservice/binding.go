package iservice

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type SvcBinding struct {
	BindingBasic
	Prices  []sdk.Coin `json:"price"`
	Levels  []int      `json:"level"`
	IsValid bool       `json:"is_valid"`
}

type BindingBasic struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
	BindingType BindingType    `json:"binding_type"`
	Deposit     sdk.Coin       `json:"deposit"`
	Expiration  int64          `json:"expiration"`
}

type Level struct {
	AvgRspTime int     `json:"avg_rsp_time"`
	UsableTime float32 `json:"usable_time"`
}

// NewSvcBinding returns a new SvcBinding with the provided values.
func NewSvcBinding(defChainID, defName, bindChainID string, provider sdk.AccAddress, bindingType BindingType, deposit sdk.Coin, prices []sdk.Coin, levels []int, expiration int64) SvcBinding {
	return SvcBinding{
		BindingBasic: BindingBasic{
			DefChainID:  defChainID,
			DefName:     defName,
			BindChainID: bindChainID,
			Provider:    provider,
			BindingType: bindingType,
			Deposit:     deposit,
			Expiration:  expiration,
		},
		Prices:  prices,
		Levels:  levels,
		IsValid: false,
	}
}

func SvcBindingEqual(bindingA, bindingB SvcBinding) bool {
	if bindingA.DefChainID == bindingB.DefChainID &&
		bindingA.DefName == bindingB.DefName &&
		bindingA.BindChainID == bindingB.BindChainID &&
		bindingA.Provider.String() == bindingB.Provider.String() &&
		bindingA.BindingType == bindingB.BindingType &&
		bindingA.Deposit.IsEqual(bindingB.Deposit) &&
		len(bindingA.Levels) == len(bindingB.Levels) &&
		len(bindingA.Prices) == len(bindingB.Prices) &&
		bindingA.Expiration == bindingB.Expiration {
		for i, level := range bindingA.Levels {
			if level != bindingB.Levels[i] {
				return false
			}
		}
		for j, prices := range bindingA.Prices {
			if !prices.IsEqual(bindingB.Prices[j]) {
				return false
			}
		}
		return true
	}
	return false
}

type BindingType byte

const (
	Global BindingType = 0x01
	Local  BindingType = 0x02
)

// String to BindingType byte, Returns ff if invalid.
func BindingTypeFromString(str string) (BindingType, error) {
	switch str {
	case "Local":
		return Local, nil
	case "Global":
		return Global, nil
	default:
		return BindingType(0xff), errors.Errorf("'%s' is not a valid binding type", str)
	}
}

// is defined BindingType?
func validBindingType(bt BindingType) bool {
	if bt == Local ||
		bt == Global {
		return true
	}
	return false
}

// For Printf / Sprintf, returns bech32 when using %s
func (bt BindingType) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", bt.String())))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(bt))))
	}
}

// Turns BindingType byte to String
func (bt BindingType) String() string {
	switch bt {
	case Local:
		return "Local"
	case Global:
		return "Global"
	default:
		return ""
	}
}

// Marshals to JSON using string
func (bt BindingType) MarshalJSON() ([]byte, error) {
	return json.Marshal(bt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (bt *BindingType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := BindingTypeFromString(s)
	if err != nil {
		return err
	}
	*bt = bz2
	return nil
}