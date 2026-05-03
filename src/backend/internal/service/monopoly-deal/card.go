package monopoly_deal

import (
	"fmt"
	"net/url"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/schema/monopoly_deal_schema"
)

var actionMoneyMap = map[monopoly_deal.AssetKey]monopoly_deal.AssetKey{
	monopoly_deal.AssetKeyDealBreaker:         monopoly_deal.AssetKeyMoney5,
	monopoly_deal.AssetKeyJustSayNo:           monopoly_deal.AssetKeyMoney4,
	monopoly_deal.AssetKeyHotel:               monopoly_deal.AssetKeyMoney4,
	monopoly_deal.AssetKeyDebtCollector:       monopoly_deal.AssetKeyMoney3,
	monopoly_deal.AssetKeyForcedDeal:          monopoly_deal.AssetKeyMoney3,
	monopoly_deal.AssetKeySlyDeal:             monopoly_deal.AssetKeyMoney3,
	monopoly_deal.AssetKeyHouse:               monopoly_deal.AssetKeyMoney3,
	monopoly_deal.AssetKeyItsMyBirthday:       monopoly_deal.AssetKeyMoney2,
	monopoly_deal.AssetKeyPassGo:              monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyDoubleTheRent:       monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyRentWild:            monopoly_deal.AssetKeyMoney3,
	monopoly_deal.AssetKeyRentBrownSky:        monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyRentPinkOrange:      monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyRentRedYellow:       monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyRentGreenBlue:       monopoly_deal.AssetKeyMoney1,
	monopoly_deal.AssetKeyRentUtilityRailroad: monopoly_deal.AssetKeyMoney1,
}

var smallCardMap = map[monopoly_deal.AssetKey]string{
	monopoly_deal.AssetKeyBrown1:              "brown",
	monopoly_deal.AssetKeyBrown2:              "brown",
	monopoly_deal.AssetKeySky1:                "sky",
	monopoly_deal.AssetKeySky2:                "sky",
	monopoly_deal.AssetKeySky3:                "sky",
	monopoly_deal.AssetKeyPink1:               "pink",
	monopoly_deal.AssetKeyPink2:               "pink",
	monopoly_deal.AssetKeyPink3:               "pink",
	monopoly_deal.AssetKeyOrange1:             "orange",
	monopoly_deal.AssetKeyOrange2:             "orange",
	monopoly_deal.AssetKeyOrange3:             "orange",
	monopoly_deal.AssetKeyRed1:                "red",
	monopoly_deal.AssetKeyRed2:                "red",
	monopoly_deal.AssetKeyRed3:                "red",
	monopoly_deal.AssetKeyYellow1:             "yellow",
	monopoly_deal.AssetKeyYellow2:             "yellow",
	monopoly_deal.AssetKeyYellow3:             "yellow",
	monopoly_deal.AssetKeyGreen1:              "green",
	monopoly_deal.AssetKeyGreen2:              "green",
	monopoly_deal.AssetKeyGreen3:              "green",
	monopoly_deal.AssetKeyBlue1:               "blue",
	monopoly_deal.AssetKeyBlue2:               "blue",
	monopoly_deal.AssetKeyUtil1:               "util",
	monopoly_deal.AssetKeyUtil2:               "util",
	monopoly_deal.AssetKeyRail1:               "rail",
	monopoly_deal.AssetKeyRail2:               "rail",
	monopoly_deal.AssetKeyRail3:               "rail",
	monopoly_deal.AssetKeyRail4:               "rail",
	monopoly_deal.AssetKeyWildBrownSky:        "brown_sky",
	monopoly_deal.AssetKeyWildSkyRailroad:     "sky_rail",
	monopoly_deal.AssetKeyWildPinkOrange:      "pink_orange",
	monopoly_deal.AssetKeyWildRedYellow:       "red_yellow",
	monopoly_deal.AssetKeyWildGreenBlue:       "green_blue",
	monopoly_deal.AssetKeyWildGreenRailroad:   "green_rail",
	monopoly_deal.AssetKeyWildUtilityRailroad: "util_rail",
	monopoly_deal.AssetKeyWildWild:            "wild",
	monopoly_deal.AssetKeyMoney10:             "10",
	monopoly_deal.AssetKeyMoney5:              "5",
	monopoly_deal.AssetKeyMoney4:              "4",
	monopoly_deal.AssetKeyMoney3:              "3",
	monopoly_deal.AssetKeyMoney2:              "2",
	monopoly_deal.AssetKeyMoney1:              "1",
}

func (c *Controller) smallCardImages() []*monopoly_deal_schema.AssetImage {
	aks := monopoly_deal.AllAssetKeys()
	ai := make([]*monopoly_deal_schema.AssetImage, 0, len(aks))
	for _, ak := range aks {
		key, ok := actionMoneyMap[ak]
		if !ok {
			key = ak
		}

		u, _ := url.JoinPath(fmt.Sprintf("https://%s", c.cfg.BackendDomain), "static", "card", "small", fmt.Sprintf("%s.svg", smallCardMap[key]))
		ai = append(ai, &monopoly_deal_schema.AssetImage{
			Kind:     monopoly_deal_schema.AssetImageKind_ASSET_IMAGE_KIND_SMALL,
			AssetKey: ak.Proto(),
			ImageUrl: u,
		})
	}
	return ai
}

var largeCardMap = map[monopoly_deal.AssetKey]string{
	monopoly_deal.AssetKeyBrown1:              "brown1",
	monopoly_deal.AssetKeyBrown2:              "brown2",
	monopoly_deal.AssetKeySky1:                "sky1",
	monopoly_deal.AssetKeySky2:                "sky2",
	monopoly_deal.AssetKeySky3:                "sky3",
	monopoly_deal.AssetKeyPink1:               "pink1",
	monopoly_deal.AssetKeyPink2:               "pink2",
	monopoly_deal.AssetKeyPink3:               "pink3",
	monopoly_deal.AssetKeyOrange1:             "orange1",
	monopoly_deal.AssetKeyOrange2:             "orange2",
	monopoly_deal.AssetKeyOrange3:             "orange3",
	monopoly_deal.AssetKeyRed1:                "red1",
	monopoly_deal.AssetKeyRed2:                "red2",
	monopoly_deal.AssetKeyRed3:                "red3",
	monopoly_deal.AssetKeyYellow1:             "yellow1",
	monopoly_deal.AssetKeyYellow2:             "yellow2",
	monopoly_deal.AssetKeyYellow3:             "yellow3",
	monopoly_deal.AssetKeyGreen1:              "green1",
	monopoly_deal.AssetKeyGreen2:              "green2",
	monopoly_deal.AssetKeyGreen3:              "green3",
	monopoly_deal.AssetKeyBlue1:               "blue1",
	monopoly_deal.AssetKeyBlue2:               "blue2",
	monopoly_deal.AssetKeyUtil1:               "util1",
	monopoly_deal.AssetKeyUtil2:               "util2",
	monopoly_deal.AssetKeyRail1:               "rail1",
	monopoly_deal.AssetKeyRail2:               "rail2",
	monopoly_deal.AssetKeyRail3:               "rail3",
	monopoly_deal.AssetKeyRail4:               "rail4",
	monopoly_deal.AssetKeyWildBrownSky:        "brown_sky",
	monopoly_deal.AssetKeyWildSkyRailroad:     "sky_rail",
	monopoly_deal.AssetKeyWildPinkOrange:      "pink_orange",
	monopoly_deal.AssetKeyWildRedYellow:       "red_yellow",
	monopoly_deal.AssetKeyWildGreenBlue:       "green_blue",
	monopoly_deal.AssetKeyWildGreenRailroad:   "green_rail",
	monopoly_deal.AssetKeyWildUtilityRailroad: "util_rail",
	monopoly_deal.AssetKeyWildWild:            "wild",
	monopoly_deal.AssetKeyMoney10:             "10",
	monopoly_deal.AssetKeyMoney5:              "5",
	monopoly_deal.AssetKeyMoney4:              "4",
	monopoly_deal.AssetKeyMoney3:              "3",
	monopoly_deal.AssetKeyMoney2:              "2",
	monopoly_deal.AssetKeyMoney1:              "1",
	monopoly_deal.AssetKeyDealBreaker:         "set_snatcher",
	monopoly_deal.AssetKeyJustSayNo:           "nah",
	monopoly_deal.AssetKeyHotel:               "hotel",
	monopoly_deal.AssetKeyDebtCollector:       "settle_up",
	monopoly_deal.AssetKeyForcedDeal:          "property_swap",
	monopoly_deal.AssetKeySlyDeal:             "property_steal",
	monopoly_deal.AssetKeyHouse:               "house",
	monopoly_deal.AssetKeyItsMyBirthday:       "party_bill",
	monopoly_deal.AssetKeyDoubleTheRent:       "double_the_rent",
	monopoly_deal.AssetKeyPassGo:              "payday",
	monopoly_deal.AssetKeyRentWild:            "rent_wild",
	monopoly_deal.AssetKeyRentBrownSky:        "rent_brown_sky",
	monopoly_deal.AssetKeyRentPinkOrange:      "rent_pink_orange",
	monopoly_deal.AssetKeyRentRedYellow:       "rent_red_yellow",
	monopoly_deal.AssetKeyRentGreenBlue:       "rent_green_blue",
	monopoly_deal.AssetKeyRentUtilityRailroad: "rent_util_rail",
}

func (c *Controller) largeCardImages() []*monopoly_deal_schema.AssetImage {
	aks := monopoly_deal.AllAssetKeys()
	ai := make([]*monopoly_deal_schema.AssetImage, 0, len(aks))
	for _, ak := range aks {
		u, _ := url.JoinPath(fmt.Sprintf("https://%s", c.cfg.BackendDomain), "static", "card", "large", fmt.Sprintf("%s.svg", largeCardMap[ak]))
		ai = append(ai, &monopoly_deal_schema.AssetImage{
			Kind:     monopoly_deal_schema.AssetImageKind_ASSET_IMAGE_KIND_LARGE,
			AssetKey: ak.Proto(),
			ImageUrl: u,
		})
	}
	return ai
}
