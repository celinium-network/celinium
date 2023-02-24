package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appparams "celinium/app/params"
	"celinium/x/tokenfactory/types"
)

func TestDeconstructDenom(t *testing.T) {
	appparams.SetAddressPrefixes()

	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            "factory/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/bitcoin",
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            "factory/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/bitcoin/1",
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            "factory/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/",
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/bitcoin",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            "factory/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/////",
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: "factory/demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0"
			creator, subdenom, err := types.DeconstructDenom(tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "demo1wxhrdslu9w2phqqx6yxqz4vk5nfejm7lf7fsr0jhgjhgkhjklhkjhkjhgjhgjgjghelugt",
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
