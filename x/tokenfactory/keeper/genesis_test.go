package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"celinium/app"
	"celinium/x/tokenfactory/types"
)

func (suite *IntegrationTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf",
				},
			},
			{
				Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
				},
			},
			{
				Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf",
				},
			},
		},
	}

	//suite.SetupTestForInitGenesis()
	app := app.SetupTestForInitGenesis(suite.T())
	suite.Ctx = app.NewContext(true, tmtypes.Header{})

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			app.BankKeeper.SetDenomMetaData(suite.Ctx, banktypes.Metadata{Base: denom.GetDenom()})
		}
	}

	// check before initGenesis that the module account is nil
	tokenfactoryModuleAccount := app.AccountKeeper.GetAccount(suite.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	suite.Require().Nil(tokenfactoryModuleAccount)

	app.TokenFactoryKeeper.SetParams(suite.Ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("uosmo", 100)}})
	app.TokenFactoryKeeper.InitGenesis(suite.Ctx, genesisState)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(suite.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	suite.Require().NotNil(tokenfactoryModuleAccount)

	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(suite.Ctx)
	suite.Require().NotNil(exportedGenesis)
	suite.Require().Equal(genesisState, *exportedGenesis)
}
