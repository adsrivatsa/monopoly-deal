package deal_no_mercy

import (
	"fmt"
	"net/url"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema/deal_no_mercy_schema"
)

// actionMoneyMap maps each action card to the money tile it shares a small
// face with (actions have no dedicated small tile — see
// docs/deal-no-mercy/card-list.md "action→small-tile mapping").
var actionMoneyMap = map[deal_no_mercy.AssetKey]deal_no_mercy.AssetKey{
	deal_no_mercy.AssetKeyBigPayday: deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyGoAgain:   deal_no_mercy.AssetKeyMoney1,

	deal_no_mercy.AssetKeyDebtTrap: deal_no_mercy.AssetKeyMoney2,
	deal_no_mercy.AssetKeyYoink:    deal_no_mercy.AssetKeyMoney2,

	deal_no_mercy.AssetKeySetSnatcher:  deal_no_mercy.AssetKeyMoney3,
	deal_no_mercy.AssetKeyShack:        deal_no_mercy.AssetKeyMoney3,
	deal_no_mercy.AssetKeyPropertyRaid: deal_no_mercy.AssetKeyMoney3,

	deal_no_mercy.AssetKeyHeist:       deal_no_mercy.AssetKeyMoney4,
	deal_no_mercy.AssetKeyMarketCrash: deal_no_mercy.AssetKeyMoney4,
	deal_no_mercy.AssetKeyBankSwap:    deal_no_mercy.AssetKeyMoney4,
	deal_no_mercy.AssetKeyNah:         deal_no_mercy.AssetKeyMoney4,

	deal_no_mercy.AssetKeyRepoMan:    deal_no_mercy.AssetKeyMoney5,
	deal_no_mercy.AssetKeyTaxDay:     deal_no_mercy.AssetKeyMoney5,
	deal_no_mercy.AssetKeyPickpocket: deal_no_mercy.AssetKeyMoney5,

	// Rent cards map to their money value (rent 3M / double rent wild 3M,
	// two-colour rents 1M, double rents 1M).
	deal_no_mercy.AssetKeyRentWild:            deal_no_mercy.AssetKeyMoney3,
	deal_no_mercy.AssetKeyDoubleRentWild:      deal_no_mercy.AssetKeyMoney3,
	deal_no_mercy.AssetKeyRentBrownSky:        deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyRentPinkOrange:      deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyRentRedYellow:       deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyRentGreenBlue:       deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyRentUtilityRailroad: deal_no_mercy.AssetKeyMoney1,

	deal_no_mercy.AssetKeyDoubleRentBrownSky:        deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyDoubleRentPinkOrange:      deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyDoubleRentRedYellow:       deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyDoubleRentGreenBlue:       deal_no_mercy.AssetKeyMoney1,
	deal_no_mercy.AssetKeyDoubleRentUtilityRailroad: deal_no_mercy.AssetKeyMoney1,
}

// smallCardMap maps a property / money asset key to its small colour/value
// tile basename under /static/deal-no-mercy/card/small.
var smallCardMap = map[deal_no_mercy.AssetKey]string{
	deal_no_mercy.AssetKeyBrown1:  "brown",
	deal_no_mercy.AssetKeyBrown2:  "brown",
	deal_no_mercy.AssetKeySky1:    "sky",
	deal_no_mercy.AssetKeySky2:    "sky",
	deal_no_mercy.AssetKeySky3:    "sky",
	deal_no_mercy.AssetKeyPink1:   "pink",
	deal_no_mercy.AssetKeyPink2:   "pink",
	deal_no_mercy.AssetKeyPink3:   "pink",
	deal_no_mercy.AssetKeyOrange1: "orange",
	deal_no_mercy.AssetKeyOrange2: "orange",
	deal_no_mercy.AssetKeyOrange3: "orange",
	deal_no_mercy.AssetKeyRed1:    "red",
	deal_no_mercy.AssetKeyRed2:    "red",
	deal_no_mercy.AssetKeyRed3:    "red",
	deal_no_mercy.AssetKeyYellow1: "yellow",
	deal_no_mercy.AssetKeyYellow2: "yellow",
	deal_no_mercy.AssetKeyYellow3: "yellow",
	deal_no_mercy.AssetKeyGreen1:  "green",
	deal_no_mercy.AssetKeyGreen2:  "green",
	deal_no_mercy.AssetKeyGreen3:  "green",
	deal_no_mercy.AssetKeyBlue1:   "blue",
	deal_no_mercy.AssetKeyBlue2:   "blue",
	deal_no_mercy.AssetKeyUtil1:   "util",
	deal_no_mercy.AssetKeyUtil2:   "util",
	deal_no_mercy.AssetKeyRail1:   "rail",
	deal_no_mercy.AssetKeyRail2:   "rail",
	deal_no_mercy.AssetKeyRail3:   "rail",
	deal_no_mercy.AssetKeyRail4:   "rail",

	deal_no_mercy.AssetKeyWildBrownSky:        "brown_sky",
	deal_no_mercy.AssetKeyWildSkyRailroad:     "sky_rail",
	deal_no_mercy.AssetKeyWildPinkOrange:      "pink_orange",
	deal_no_mercy.AssetKeyWildRedYellow:       "red_yellow",
	deal_no_mercy.AssetKeyWildGreenBlue:       "green_blue",
	deal_no_mercy.AssetKeyWildGreenRailroad:   "green_rail",
	deal_no_mercy.AssetKeyWildUtilityRailroad: "util_rail",
	deal_no_mercy.AssetKeyWildWild:            "wild",

	deal_no_mercy.AssetKeyMoney15: "15",
	deal_no_mercy.AssetKeyMoney10: "10",
	deal_no_mercy.AssetKeyMoney5:  "5",
	deal_no_mercy.AssetKeyMoney4:  "4",
	deal_no_mercy.AssetKeyMoney3:  "3",
	deal_no_mercy.AssetKeyMoney2:  "2",
	deal_no_mercy.AssetKeyMoney1:  "1",
}

func (c *Controller) smallCardImages() []*deal_no_mercy_schema.AssetImage {
	aks := deal_no_mercy.AllAssetKeys()
	ai := make([]*deal_no_mercy_schema.AssetImage, 0, len(aks))
	for _, ak := range aks {
		key, ok := actionMoneyMap[ak]
		if !ok {
			key = ak
		}

		u, _ := url.JoinPath(fmt.Sprintf("https://%s", c.cfg.BackendDomain), "static", "deal-no-mercy", "card", "small", fmt.Sprintf("%s.svg", smallCardMap[key]))
		ai = append(ai, &deal_no_mercy_schema.AssetImage{
			Kind:     deal_no_mercy_schema.AssetImageKind_ASSET_IMAGE_KIND_SMALL,
			AssetKey: ak.Proto(),
			ImageUrl: u,
		})
	}
	return ai
}

// largeCardMap maps every asset key to its large face basename under
// /static/deal-no-mercy/card/large.
var largeCardMap = map[deal_no_mercy.AssetKey]string{
	deal_no_mercy.AssetKeyBrown1:  "brown1",
	deal_no_mercy.AssetKeyBrown2:  "brown2",
	deal_no_mercy.AssetKeySky1:    "sky1",
	deal_no_mercy.AssetKeySky2:    "sky2",
	deal_no_mercy.AssetKeySky3:    "sky3",
	deal_no_mercy.AssetKeyPink1:   "pink1",
	deal_no_mercy.AssetKeyPink2:   "pink2",
	deal_no_mercy.AssetKeyPink3:   "pink3",
	deal_no_mercy.AssetKeyOrange1: "orange1",
	deal_no_mercy.AssetKeyOrange2: "orange2",
	deal_no_mercy.AssetKeyOrange3: "orange3",
	deal_no_mercy.AssetKeyRed1:    "red1",
	deal_no_mercy.AssetKeyRed2:    "red2",
	deal_no_mercy.AssetKeyRed3:    "red3",
	deal_no_mercy.AssetKeyYellow1: "yellow1",
	deal_no_mercy.AssetKeyYellow2: "yellow2",
	deal_no_mercy.AssetKeyYellow3: "yellow3",
	deal_no_mercy.AssetKeyGreen1:  "green1",
	deal_no_mercy.AssetKeyGreen2:  "green2",
	deal_no_mercy.AssetKeyGreen3:  "green3",
	deal_no_mercy.AssetKeyBlue1:   "blue1",
	deal_no_mercy.AssetKeyBlue2:   "blue2",
	deal_no_mercy.AssetKeyUtil1:   "util1",
	deal_no_mercy.AssetKeyUtil2:   "util2",
	deal_no_mercy.AssetKeyRail1:   "rail1",
	deal_no_mercy.AssetKeyRail2:   "rail2",
	deal_no_mercy.AssetKeyRail3:   "rail3",
	deal_no_mercy.AssetKeyRail4:   "rail4",

	deal_no_mercy.AssetKeyWildBrownSky:        "brown_sky",
	deal_no_mercy.AssetKeyWildSkyRailroad:     "sky_rail",
	deal_no_mercy.AssetKeyWildPinkOrange:      "pink_orange",
	deal_no_mercy.AssetKeyWildRedYellow:       "red_yellow",
	deal_no_mercy.AssetKeyWildGreenBlue:       "green_blue",
	deal_no_mercy.AssetKeyWildGreenRailroad:   "green_rail",
	deal_no_mercy.AssetKeyWildUtilityRailroad: "util_rail",
	deal_no_mercy.AssetKeyWildWild:            "wild",

	deal_no_mercy.AssetKeyMoney15: "15",
	deal_no_mercy.AssetKeyMoney10: "10",
	deal_no_mercy.AssetKeyMoney5:  "5",
	deal_no_mercy.AssetKeyMoney4:  "4",
	deal_no_mercy.AssetKeyMoney3:  "3",
	deal_no_mercy.AssetKeyMoney2:  "2",
	deal_no_mercy.AssetKeyMoney1:  "1",

	deal_no_mercy.AssetKeySetSnatcher:  "set_snatcher",
	deal_no_mercy.AssetKeyDebtTrap:     "debt_trap",
	deal_no_mercy.AssetKeyGoAgain:      "go_again",
	deal_no_mercy.AssetKeyHeist:        "heist",
	deal_no_mercy.AssetKeyMarketCrash:  "market_crash",
	deal_no_mercy.AssetKeyBigPayday:    "big_payday",
	deal_no_mercy.AssetKeyRepoMan:      "repo_man",
	deal_no_mercy.AssetKeyShack:        "shack",
	deal_no_mercy.AssetKeyPropertyRaid: "property_raid",
	deal_no_mercy.AssetKeyTaxDay:       "tax_day",
	deal_no_mercy.AssetKeyPickpocket:   "pickpocket",
	deal_no_mercy.AssetKeyBankSwap:     "bank_swap",
	deal_no_mercy.AssetKeyYoink:        "yoink",
	deal_no_mercy.AssetKeyNah:          "nah",

	deal_no_mercy.AssetKeyRentWild:            "rent_wild",
	deal_no_mercy.AssetKeyRentBrownSky:        "rent_brown_sky",
	deal_no_mercy.AssetKeyRentPinkOrange:      "rent_pink_orange",
	deal_no_mercy.AssetKeyRentRedYellow:       "rent_red_yellow",
	deal_no_mercy.AssetKeyRentGreenBlue:       "rent_green_blue",
	deal_no_mercy.AssetKeyRentUtilityRailroad: "rent_util_rail",

	deal_no_mercy.AssetKeyDoubleRentWild:            "double_rent_wild",
	deal_no_mercy.AssetKeyDoubleRentBrownSky:        "double_rent_brown_sky",
	deal_no_mercy.AssetKeyDoubleRentPinkOrange:      "double_rent_pink_orange",
	deal_no_mercy.AssetKeyDoubleRentRedYellow:       "double_rent_red_yellow",
	deal_no_mercy.AssetKeyDoubleRentGreenBlue:       "double_rent_green_blue",
	deal_no_mercy.AssetKeyDoubleRentUtilityRailroad: "double_rent_util_rail",
}

func (c *Controller) largeCardImages() []*deal_no_mercy_schema.AssetImage {
	aks := deal_no_mercy.AllAssetKeys()
	ai := make([]*deal_no_mercy_schema.AssetImage, 0, len(aks))
	for _, ak := range aks {
		u, _ := url.JoinPath(fmt.Sprintf("https://%s", c.cfg.BackendDomain), "static", "deal-no-mercy", "card", "large", fmt.Sprintf("%s.svg", largeCardMap[ak]))
		ai = append(ai, &deal_no_mercy_schema.AssetImage{
			Kind:     deal_no_mercy_schema.AssetImageKind_ASSET_IMAGE_KIND_LARGE,
			AssetKey: ak.Proto(),
			ImageUrl: u,
		})
	}
	return ai
}
