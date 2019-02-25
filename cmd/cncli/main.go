package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	"commercio-network"

	// CommercioAUTH
	authclient "commercio-network/x/commercioauth/client"
	authrest "commercio-network/x/commercioauth/client/rest"

	// CommercioDOCS
	docsclient "commercio-network/x/commerciodocs/client"
	docsrest "commercio-network/x/commerciodocs/client/rest"

	// CommercioID
	idclient "commercio-network/x/commercioid/client"
	idrest "commercio-network/x/commercioid/client/rest"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
)

const (
	storeAcc  = "acc"
	storeAUTH = "commercioauth"
	storeID   = "commercioid"
	storeDOCS = "commerciodocs"
)

var defaultCLIHome = os.ExpandEnv("$HOME/.cncli")

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	mc := []sdk.ModuleClients{
		// CommercioAUTH
		authclient.NewModuleClient(storeAUTH, cdc),

		// CommercioID
		idclient.NewModuleClient(storeID, cdc),

		// CommercioDOCS
		docsclient.NewModuleClient(storeDOCS, cdc),
	}

	rootCmd := &cobra.Command{
		Use:   "cncli",
		Short: "Commercio.network client",
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(defaultCLIHome),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	executor := cli.PrepareMainCmd(rootCmd, "NS", defaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	keys.RegisterRoutes(rs.Mux, rs.CliCtx.Indent)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeAcc)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)

	// CommercioAUTH
	authrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeAUTH)

	// CommercioID
	idrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeID)

	// CommercioDOCS
	docsrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeDOCS)
}

func queryCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(storeAcc, cdc),
	)

	for _, m := range mc {
		queryCmd.AddCommand(m.GetQueryCmd())
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}
