package interstaking

import (
	"encoding/json"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"celinium/x/inter-staking/keeper"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModuleBasic struct{}

// DefaultGenesis implements module.AppModuleBasic
func (AppModuleBasic) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	panic("unimplemented")
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	panic("unimplemented")
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	panic("unimplemented")
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	panic("unimplemented")
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
	panic("unimplemented")
}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(codectypes.InterfaceRegistry) {
	panic("unimplemented")
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {
	panic("unimplemented")
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	panic("unimplemented")
}

type AppModule struct {
	AppModuleBasic
	keeper.Keeper
}

// GenerateGenesisState implements module.AppModuleSimulation
func (AppModule) GenerateGenesisState(input *module.SimulationState) {
	panic("unimplemented")
}

// ProposalContents implements module.AppModuleSimulation
func (AppModule) ProposalContents(simState module.SimulationState) []simulation.WeightedProposalContent {
	panic("unimplemented")
}

// RandomizedParams implements module.AppModuleSimulation
func (AppModule) RandomizedParams(r *rand.Rand) []simulation.ParamChange {
	panic("unimplemented")
}

// RegisterStoreDecoder implements module.AppModuleSimulation
func (AppModule) RegisterStoreDecoder(sdk.StoreDecoderRegistry) {
	panic("unimplemented")
}

// WeightedOperations implements module.AppModuleSimulation
func (AppModule) WeightedOperations(simState module.SimulationState) []simulation.WeightedOperation {
	panic("unimplemented")
}

// DefaultGenesis implements module.AppModule
func (AppModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	panic("unimplemented")
}

// GetQueryCmd implements module.AppModule
func (AppModule) GetQueryCmd() *cobra.Command {
	panic("unimplemented")
}

// GetTxCmd implements module.AppModule
func (AppModule) GetTxCmd() *cobra.Command {
	panic("unimplemented")
}

// Name implements module.AppModule
func (AppModule) Name() string {
	panic("unimplemented")
}

// RegisterGRPCGatewayRoutes implements module.AppModule
func (AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
	panic("unimplemented")
}

// RegisterInterfaces implements module.AppModule
func (AppModule) RegisterInterfaces(codectypes.InterfaceRegistry) {
	panic("unimplemented")
}

// RegisterLegacyAminoCodec implements module.AppModule
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {
	panic("unimplemented")
}

// ValidateGenesis implements module.AppModule
func (AppModule) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	panic("unimplemented")
}

// ExportGenesis implements module.AppModule
func (AppModule) ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage {
	panic("unimplemented")
}

// InitGenesis implements module.AppModule
func (AppModule) InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate {
	panic("unimplemented")
}

// ConsensusVersion implements module.AppModule
func (AppModule) ConsensusVersion() uint64 {
	panic("unimplemented")
}

// LegacyQuerierHandler implements module.AppModule
func (AppModule) LegacyQuerierHandler(*codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	panic("unimplemented")
}

// QuerierRoute implements module.AppModule
func (AppModule) QuerierRoute() string {
	panic("unimplemented")
}

// RegisterInvariants implements module.AppModule
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {
	panic("unimplemented")
}

// RegisterServices implements module.AppModule
func (AppModule) RegisterServices(module.Configurator) {
	panic("unimplemented")
}

// Route implements module.AppModule
func (AppModule) Route() sdk.Route {
	panic("unimplemented")
}
