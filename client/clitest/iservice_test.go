package clitest

import (
	"fmt"
	"os"
	"testing"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/irisnet/irishub/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/server"
)

func init() {
	irisHome, iriscliHome = getTestingHomeDirs()
}

func TestIrisCLIIserviceDefine(t *testing.T) {
	tests.ExecuteT(t, fmt.Sprintf("iris --home=%s unsafe_reset_all", irisHome), "")
	executeWrite(t, fmt.Sprintf("iriscli keys delete --home=%s foo", iriscliHome), app.DefaultKeyPass)
	executeWrite(t, fmt.Sprintf("iriscli keys delete --home=%s bar", iriscliHome), app.DefaultKeyPass)
	chainID, _ := executeInit(t, fmt.Sprintf("iris init -o --name=foo --home=%s --home-client=%s", irisHome, iriscliHome))
	executeWrite(t, fmt.Sprintf("iriscli keys add --home=%s bar", iriscliHome), app.DefaultKeyPass)

	err := modifyGenesisFileForIService(irisHome)
	require.NoError(t, err)

	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", iriscliHome, servAddr, chainID)

	// start iris server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("iris start --home=%s --rpc.laddr=%v", irisHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("iriscli keys show foo --output=json --home=%s", iriscliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("iriscli keys show bar --output=json --home=%s", iriscliHome))

	serviceName := "testService"

	serviceQuery := tests.ExecuteT(t, fmt.Sprintf("iriscli iservice definition --service-name=%s --def-chain-id=%s %v", serviceName, chainID, flags), "")
	require.Equal(t, "", serviceQuery)

	fooAcc := executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin := convertToIrisBaseAccount(t, fooAcc)
	num := getAmountFromCoinStr(fooCoin)
	require.Equal(t, "2100iris", fooCoin)

	// iservice define
	fileName := iriscliHome + string(os.PathSeparator) + "test.proto"
	defer tests.ExecuteT(t, fmt.Sprintf("rm -f %s", fileName), "")
	ioutil.WriteFile(fileName, []byte(idlContent), 0644)
	sdStr := fmt.Sprintf("iriscli iservice define %v", flags)
	sdStr += fmt.Sprintf(" --from=%s", "foo")
	sdStr += fmt.Sprintf(" --service-name=%s", serviceName)
	sdStr += fmt.Sprintf(" --service-description=%s", "test")
	sdStr += fmt.Sprintf(" --tags=%s", "tag1,tag2")
	sdStr += fmt.Sprintf(" --author-description=%s", "foo")
	sdStr += fmt.Sprintf(" --messaging=%s", "Multicast")
	sdStr += fmt.Sprintf(" --file=%s", fileName)
	sdStr += fmt.Sprintf(" --fee=%s", "0.004iris")

	executeWrite(t, sdStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin = convertToIrisBaseAccount(t, fooAcc)
	num = getAmountFromCoinStr(fooCoin)

	if !(num > 2099 && num < 2100) {
		t.Error("Test Failed: (2099, 2100) expected, recieved: {}", num)
	}

	serviceDef := executeGetServiceDefinition(t, fmt.Sprintf("iriscli iservice definition --service-name=%s --def-chain-id=%s %v", serviceName, chainID, flags))
	require.Equal(t, serviceName, serviceDef.Name)

	// method test
	require.Equal(t, "SayHello", serviceDef.Methods[0].Name)
	require.Equal(t, "sayHello", serviceDef.Methods[0].Description)
	require.Equal(t, "NoCached", serviceDef.Methods[0].OutputCached.String())
	require.Equal(t, "NoPrivacy", serviceDef.Methods[0].OutputPrivacy.String())

	// binding test
	sdStr = fmt.Sprintf("iriscli iservice bind %v", flags)
	sdStr += fmt.Sprintf(" --service-name=%s", serviceName)
	sdStr += fmt.Sprintf(" --def-chain-id=%s", chainID)
	sdStr += fmt.Sprintf(" --bind-type=%s", "Local")
	sdStr += fmt.Sprintf(" --deposit=%s", "1000iris")
	sdStr += fmt.Sprintf(" --prices=%s", "1iris")
	sdStr += fmt.Sprintf(" --avg-rsp-time=%d", 10000)
	sdStr += fmt.Sprintf(" --usable-time=%d", 10000)
	sdStr += fmt.Sprintf(" --expiration=%d", -1)
	sdStr += fmt.Sprintf(" --fee=%s", "0.004iris")

	sdStrFoo := sdStr + fmt.Sprintf(" --from=%s", "foo")
	sdStrBar := sdStr + fmt.Sprintf(" --from=%s", "bar")

	executeWrite(t, sdStrFoo, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin = convertToIrisBaseAccount(t, fooAcc)
	num = getAmountFromCoinStr(fooCoin)

	if !(num > 1099 && num < 1100) {
		t.Error("Test Failed: (1099, 1100) expected, recieved: {}", num)
	}

	executeWrite(t, fmt.Sprintf("iriscli bank send --to=%s --from=%s --amount=1050iris --fee=0.004iris %v", barAddr.String(), "foo", flags), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)
	executeWrite(t, sdStrBar, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)
	barAcc := executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", barAddr, flags))
	barCoin := convertToIrisBaseAccount(t, barAcc)
	barNum := getAmountFromCoinStr(barCoin)

	if !(barNum > 49 && barNum < 50) {
		t.Error("Test Failed: (49, 50) expected, recieved: {}", barNum)
	}

	serviceBinding := executeGetServiceBinding(t, fmt.Sprintf("iriscli iservice binding --service-name=%s --def-chain-id=%s --bind-chain-id=%s --provider=%s %v", serviceName, chainID, chainID, fooAddr.String(), flags))
	require.NotNil(t, serviceBinding)

	serviceBindings := executeGetServiceBindings(t, fmt.Sprintf("iriscli iservice bindings --service-name=%s --def-chain-id=%s %v", serviceName, chainID, flags))
	require.Equal(t, 2, len(serviceBindings))

	// binding update test
	sdStr = fmt.Sprintf("iriscli iservice update-binding %v", flags)
	sdStr += fmt.Sprintf(" --service-name=%s", serviceName)
	sdStr += fmt.Sprintf(" --def-chain-id=%s", chainID)
	sdStr += fmt.Sprintf(" --bind-type=%s", "Global")
	sdStr += fmt.Sprintf(" --deposit=%s", "10iris")
	sdStr += fmt.Sprintf(" --prices=%s", "5iris")
	sdStr += fmt.Sprintf(" --avg-rsp-time=%d", 99)
	sdStr += fmt.Sprintf(" --usable-time=%d", 99)
	sdStr += fmt.Sprintf(" --expiration=%d", 99)
	sdStr += fmt.Sprintf(" --fee=%s", "0.004iris")
	sdStr += fmt.Sprintf(" --from=%s", "bar")
	executeWrite(t, sdStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)
	barAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", barAddr, flags))
	barCoin = convertToIrisBaseAccount(t, barAcc)
	barNum = getAmountFromCoinStr(barCoin)

	if !(barNum > 39 && barNum < 40) {
		t.Error("Test Failed: (39, 40) expected, recieved: {}", barNum)
	}
	serviceBindings = executeGetServiceBindings(t, fmt.Sprintf("iriscli iservice bindings --service-name=%s --def-chain-id=%s %v", serviceName, chainID, flags))
	var totalDeposit sdk.Coins
	for _, bind := range serviceBindings {
		totalDeposit = totalDeposit.Plus(bind.Deposit)
	}
	require.Equal(t, "2010000000000000000000iris-atto", totalDeposit.String())
}

const idlContent = `
	syntax = "proto3";

	// The greeting service definition.
	service Greeter {
		//@Attribute description:sayHello
		//@Attribute output_privacy:NoPrivacy
		//@Attribute output_cached:NoCached
		rpc SayHello (HelloRequest) returns (HelloReply) {}
	}

	// The request message containing the user's name.
	message HelloRequest {
		string name = 1;
	}

	// The response message containing the greetings
	message HelloReply {
		string message = 1;
	}`
