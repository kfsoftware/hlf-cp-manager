package resolvers

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	appconfig "github.com/kfsoftware/hlf-cp-manager/config"
	"github.com/kfsoftware/hlf-cp-manager/gql"
)

type Resolver struct {
	DCS                  map[string]*appconfig.DCClient
	ChannelManagerConfig appconfig.HLFChannelManagerConfig
	FabricSDK            *fabsdk.FabricSDK
}

// Mutation returns gql.MutationResolver implementation.
func (r *Resolver) Mutation() gql.MutationResolver { return &mutationResolver{r} }

// Query returns gql.QueryResolver implementation.
func (r *Resolver) Query() gql.QueryResolver { return &queryResolver{r} }

type mutationResolver struct {
	*Resolver
}

type queryResolver struct{ *Resolver }
