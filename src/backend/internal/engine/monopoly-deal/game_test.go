package monopoly_deal

import (
	"testing"

	"github.com/google/uuid"
)

// moveCardFromDeckToHand pulls the first deck card matching one of the
// categories into the player's hand so tests can play it deterministically.
func moveCardFromDeckToHand(t *testing.T, g *Game, playerID uuid.UUID, categories ...Category) Card {
	t.Helper()

	for i, c := range g.Deck.Cards {
		for _, cat := range categories {
			if c.Category != cat {
				continue
			}

			g.Deck.Cards = append(g.Deck.Cards[:i], g.Deck.Cards[i+1:]...)
			hand := g.Hands[playerID]
			hand.Add(c)
			g.Hands[playerID] = hand
			return c
		}
	}

	t.Fatalf("no card of categories %v left in deck", categories)
	return Card{}
}

// assertNoDuplicateCardIDs walks every zone (deck, hands, money, properties)
// and fails if any card id appears more than once.
func assertNoDuplicateCardIDs(t *testing.T, g *Game) {
	t.Helper()

	seen := make(map[Identifier]string)
	record := func(zone string, cards Cards) {
		for _, c := range cards {
			if prev, ok := seen[c.ID]; ok {
				t.Fatalf("card %s (%s) exists in both %s and %s", c.ID, c.AssetKey, prev, zone)
			}
			seen[c.ID] = zone
		}
	}

	record("deck", g.Deck.Cards)
	for pid, hand := range g.Hands {
		record("hand:"+pid.String(), hand)
	}
	for pid, money := range g.Money {
		record("money:"+pid.String(), money)
	}
	for pid, sets := range g.Properties {
		for _, set := range sets {
			record("properties:"+pid.String(), set.Cards)
		}
	}
}

func TestPlayMoneyDoesNotDuplicateCard(t *testing.T) {
	cfg := DefaultSettings()
	players := []uuid.UUID{uuid.New(), uuid.New()}
	g := NewGame(cfg, players)
	pid := g.Players[g.CurrPlayerIdx]

	card := moveCardFromDeckToHand(t, g, pid, CategoryMoney)

	_, err := g.PlayMoney(pid, card.ID)
	if err != nil {
		t.Fatal(err)
	}

	assertNoDuplicateCardIDs(t, g)
}

func TestPlayPropertyDoesNotDuplicateCard(t *testing.T) {
	cfg := DefaultSettings()
	players := []uuid.UUID{uuid.New(), uuid.New()}
	g := NewGame(cfg, players)
	pid := g.Players[g.CurrPlayerIdx]

	card := moveCardFromDeckToHand(t, g, pid, CategoryPureProperty)

	_, err := g.PlayProperty(pid, card.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertNoDuplicateCardIDs(t, g)
}

func TestPlayHouseDoesNotDuplicateCard(t *testing.T) {
	cfg := DefaultSettings()
	players := []uuid.UUID{uuid.New(), uuid.New()}
	g := NewGame(cfg, players)
	pid := g.Players[g.CurrPlayerIdx]

	// build a complete brown set (2 cards) directly
	setID := g.IDGenerator.New()
	set := NewPropertySet(setID, ColorBrown)
	for i := 0; i < CompleteSetSize[ColorBrown]; i++ {
		var card Card
		for j, c := range g.Deck.Cards {
			if c.Category == CategoryPureProperty && c.ActiveColor == ColorBrown {
				card = c
				g.Deck.Cards = append(g.Deck.Cards[:j], g.Deck.Cards[j+1:]...)
				break
			}
		}
		if card.ID == "" {
			t.Skip("not enough brown properties in deck")
		}
		set.Cards.Add(card)
	}
	props := g.Properties[pid]
	props.Add(set)
	g.Properties[pid] = props

	houseCard := Card{}
	for j, c := range g.Deck.Cards {
		if c.AssetKey == AssetKeyHouse {
			houseCard = c
			g.Deck.Cards = append(g.Deck.Cards[:j], g.Deck.Cards[j+1:]...)
			break
		}
	}
	if houseCard.ID == "" {
		t.Skip("no house card in deck")
	}
	hand := g.Hands[pid]
	hand.Add(houseCard)
	g.Hands[pid] = hand

	_, err := g.PlayHouse(pid, houseCard.ID, setID)
	if err != nil {
		t.Fatal(err)
	}

	assertNoDuplicateCardIDs(t, g)
}
