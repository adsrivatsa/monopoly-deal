import { decode, encode } from "@msgpack/msgpack";

type EncodedSettings = string | Uint8Array | number[];

export type ShortPlayer = {
  id: string;
  name: string;
  imageUrl: string;
  isHost: boolean;
  isReady: boolean;
};

export enum Game {
  MonopolyDeal = "monopoly_deal",
  DealNoMercy = "deal_no_mercy",
}

export const supportedGames: Game[] = [Game.MonopolyDeal, Game.DealNoMercy];

export type MonopolyDealSettings = {
  // deal rules
  num_decks: number;
  start_num_cards: number;
  max_hand_size: number;
  moves_per_turn: number;

  // card rules
  payday_draw: number;
  party_bill_payment: number;
  settle_up_payment: number;

  // win conditions
  win_set_amount: number;
  win_money_amount: number;

  // game speed
  speed: number;

  // deck rules
  payday_amount: number;
  double_the_rent_amount: number;
  party_bill_amount: number;
  house_amount: number;
  property_steal_amount: number;
  property_swap_amount: number;
  settle_up_amount: number;
  hotel_amount: number;
  nah_amount: number;
  set_snatcher_amount: number;
  rent_amount: number;
  wild_rent_amount: number;

  // nah! rules
  nah_consumes_move: boolean;
};

// DealNoMercySettings mirrors the msgpack tags in
// src/backend/internal/engine/deal-no-mercy/settings.go EXACTLY (snake_case).
// The derived timeout fields (move_timeout / demand_timeout) are computed on
// the backend from speed and are intentionally omitted here.
export type DealNoMercySettings = {
  num_decks: number;

  // deal rules
  start_num_cards: number;
  max_hand_size: number;
  moves_per_turn: number;
  turn_draw: number;

  // win conditions
  win_set_amount: number;
  win_money_amount: number;

  // no mercy rules
  debt_chips_per_player: number;
  yoink_payment: number;
  shack_rent_bonus: number;
  big_payday_hand_target: number;

  // game speed
  speed: number;

  // deck rules
  set_snatcher_amount: number;
  debt_trap_amount: number;
  go_again_amount: number;
  heist_amount: number;
  market_crash_amount: number;
  big_payday_amount: number;
  repo_man_amount: number;
  shack_amount: number;
  property_raid_amount: number;
  tax_day_amount: number;
  pickpocket_amount: number;
  bank_swap_amount: number;
  yoink_amount: number;
  nah_amount: number;
  rent_amount: number;
  double_rent_amount: number;
  wild_rent_amount: number;
  double_rent_wild_amount: number;

  // nah! rules
  nah_consumes_move: boolean;
};

type GameSettingsByGame = {
  [Game.MonopolyDeal]: MonopolyDealSettings;
  [Game.DealNoMercy]: DealNoMercySettings;
};

export type GameSettingsFor<TGame extends Game> = GameSettingsByGame[TGame];

type NumericGameSettingDefinition = {
  key: string;
  label: string;
  min: number;
  max: number;
  group: string;
  kind?: "number";
};

type BooleanGameSettingDefinition = {
  key: string;
  label: string;
  group: string;
  kind: "boolean";
  options: Array<{ value: string; label: string }>;
};

export type GameSettingDefinition =
  | NumericGameSettingDefinition
  | BooleanGameSettingDefinition;

export type GameSettingSelectValue = {
  key: string;
  label: string;
  value: string;
  group: string;
  kind: "number" | "boolean";
  options: Array<{ value: string; label: string }>;
};

export type CapacityRange = {
  min: number;
  max: number;
};

const assertNever = (_value: never): never => {
  throw new Error("Unsupported game");
};

const gameNames: Record<Game, string> = {
  [Game.MonopolyDeal]: "Monopoly Deal",
  [Game.DealNoMercy]: "Deal No Mercy",
};

export const parseGame = (gameKey: string): Game | null => {
  switch (gameKey) {
    case Game.MonopolyDeal:
      return Game.MonopolyDeal;
    case Game.DealNoMercy:
      return Game.DealNoMercy;
    default:
      return null;
  }
};

export const getGameDisplayName = (gameKey: string): string => {
  const game = parseGame(gameKey);
  return game ? gameNames[game] : "Unknown Game";
};

export const getDefaultSettingsForGame = <TGame extends Game>(
  game: TGame,
): GameSettingsFor<TGame> => {
  switch (game) {
    case Game.MonopolyDeal:
      return {
        // deal rules
        num_decks: 1,
        start_num_cards: 5,
        max_hand_size: 7,
        moves_per_turn: 3,
        // card rules
        payday_draw: 2,
        party_bill_payment: 2,
        settle_up_payment: 5,
        // win conditions
        win_set_amount: 3,
        win_money_amount: 0,
        // game speed
        speed: 2,
        // deck rules
        payday_amount: 10,
        double_the_rent_amount: 2,
        party_bill_amount: 3,
        house_amount: 3,
        property_steal_amount: 3,
        property_swap_amount: 3,
        settle_up_amount: 3,
        hotel_amount: 2,
        nah_amount: 3,
        set_snatcher_amount: 2,
        rent_amount: 2,
        wild_rent_amount: 3,
        // nah! rules
        nah_consumes_move: true,
      } as GameSettingsFor<TGame>;
    case Game.DealNoMercy:
      return {
        num_decks: 1,
        // deal rules
        start_num_cards: 5,
        max_hand_size: 7,
        moves_per_turn: 3,
        turn_draw: 2,
        // win conditions
        win_set_amount: 3,
        win_money_amount: 0,
        // no mercy rules
        debt_chips_per_player: 3,
        yoink_payment: 10,
        shack_rent_bonus: 5,
        big_payday_hand_target: 7,
        // game speed
        speed: 2,
        // deck rules
        set_snatcher_amount: 3,
        debt_trap_amount: 3,
        go_again_amount: 3,
        heist_amount: 3,
        market_crash_amount: 3,
        big_payday_amount: 4,
        repo_man_amount: 2,
        shack_amount: 3,
        property_raid_amount: 3,
        tax_day_amount: 2,
        pickpocket_amount: 3,
        bank_swap_amount: 2,
        yoink_amount: 3,
        nah_amount: 5,
        rent_amount: 1,
        double_rent_amount: 1,
        wild_rent_amount: 3,
        double_rent_wild_amount: 3,
        // nah! rules
        nah_consumes_move: true,
      } as GameSettingsFor<TGame>;
    default:
      return assertNever(game as never);
  }
};

export const stringifyGameSettings = <TGame extends Game>(
  _game: TGame,
  settings: GameSettingsFor<TGame>,
): Uint8Array => {
  return encode(settings as Record<string, unknown>);
};

const buildRangeOptions = (
  min: number,
  max: number,
): Array<{ value: string; label: string }> => {
  return Array.from({ length: max - min + 1 }, (_, index) => {
    const value = String(min + index);
    return { value, label: value };
  });
};

const buildSpeedOptions = (): Array<{ value: string; label: string }> => {
  return [
    { value: "1", label: "Slow" },
    { value: "2", label: "Medium" },
    { value: "3", label: "Fast" },
  ];
};

const inRange = (value: number, min: number, max: number): boolean => {
  return Number.isInteger(value) && value >= min && value <= max;
};

const decodeMsgPackObject = (
  settings: Uint8Array,
): Record<string, unknown> | null => {
  try {
    const parsed = decode(settings) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null;
    }
    return parsed as Record<string, unknown>;
  } catch {
    return null;
  }
};

const decodeBase64 = (value: string): Uint8Array | null => {
  try {
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    const paddingLength = normalized.length % 4;
    const padded =
      paddingLength === 0
        ? normalized
        : normalized.padEnd(normalized.length + (4 - paddingLength), "=");
    const decoded = globalThis.atob(padded);
    const bytes = new Uint8Array(decoded.length);
    for (let index = 0; index < decoded.length; index += 1) {
      bytes[index] = decoded.charCodeAt(index);
    }
    return bytes;
  } catch {
    return null;
  }
};

const toUint8Array = (value: number[]): Uint8Array | null => {
  if (
    value.some((entry) => !Number.isInteger(entry) || entry < 0 || entry > 255)
  ) {
    return null;
  }
  return new Uint8Array(value);
};

const getSettingDefinitionsForGame = (game: Game): GameSettingDefinition[] => {
  switch (game) {
    case Game.MonopolyDeal:
      return [
        // speed (always first via sort)
        {
          key: "speed",
          label: "Speed",
          min: 1,
          max: 3,
          group: "Game Speed",
        },
        // deal rules
        {
          key: "num_decks",
          label: "Number of decks",
          min: 1,
          max: 3,
          group: "Deal Rules",
        },
        {
          key: "start_num_cards",
          label: "Starting cards",
          min: 5,
          max: 8,
          group: "Deal Rules",
        },
        {
          key: "max_hand_size",
          label: "Max hand size",
          min: 5,
          max: 10,
          group: "Deal Rules",
        },
        {
          key: "moves_per_turn",
          label: "Moves per turn",
          min: 3,
          max: 5,
          group: "Deal Rules",
        },
        // card rules
        {
          key: "payday_draw",
          label: "Payday draw",
          min: 2,
          max: 5,
          group: "Card Rules",
        },
        {
          key: "party_bill_payment",
          label: "Party bill payment",
          min: 2,
          max: 5,
          group: "Card Rules",
        },
        {
          key: "settle_up_payment",
          label: "Settle up payment",
          min: 5,
          max: 8,
          group: "Card Rules",
        },
        // win conditions
        {
          key: "win_set_amount",
          label: "Win set amount",
          min: 3,
          max: 6,
          group: "Win Conditions",
        },
        {
          key: "win_money_amount",
          label: "Win money amount",
          min: 0,
          max: 40,
          group: "Win Conditions",
        },
        // deck rules
        {
          key: "payday_amount",
          label: "Payday cards",
          min: 10,
          max: 20,
          group: "Deck Rules",
        },
        {
          key: "double_the_rent_amount",
          label: "Double the Rent cards",
          min: 2,
          max: 4,
          group: "Deck Rules",
        },
        {
          key: "party_bill_amount",
          label: "Party Bill cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "house_amount",
          label: "House cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "property_steal_amount",
          label: "Property Steal cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "property_swap_amount",
          label: "Property Swap cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "settle_up_amount",
          label: "Settle Up cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "hotel_amount",
          label: "Hotel cards",
          min: 2,
          max: 4,
          group: "Deck Rules",
        },
        {
          key: "nah_amount",
          label: "Nah cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        {
          key: "set_snatcher_amount",
          label: "Set Snatcher cards",
          min: 2,
          max: 4,
          group: "Deck Rules",
        },
        {
          key: "rent_amount",
          label: "Rent cards",
          min: 2,
          max: 4,
          group: "Deck Rules",
        },
        {
          key: "wild_rent_amount",
          label: "Wild Rent cards",
          min: 3,
          max: 6,
          group: "Deck Rules",
        },
        // nah! rules
        {
          key: "nah_consumes_move",
          label: "Nah consumes move",
          group: "Nah! Rules",
          kind: "boolean",
          options: [
            { value: "true", label: "On" },
            { value: "false", label: "Off" },
          ],
        },
      ];
    case Game.DealNoMercy:
      return [
        // speed (always first via sort)
        {
          key: "speed",
          label: "Speed",
          min: 1,
          max: 3,
          group: "Game Speed",
        },
        // deal rules
        {
          key: "num_decks",
          label: "Number of decks",
          min: 1,
          max: 3,
          group: "Deal Rules",
        },
        {
          key: "start_num_cards",
          label: "Starting cards",
          min: 5,
          max: 8,
          group: "Deal Rules",
        },
        {
          key: "max_hand_size",
          label: "Max hand size",
          min: 5,
          max: 10,
          group: "Deal Rules",
        },
        {
          key: "moves_per_turn",
          label: "Moves per turn",
          min: 3,
          max: 5,
          group: "Deal Rules",
        },
        {
          key: "turn_draw",
          label: "Cards drawn per turn",
          min: 2,
          max: 5,
          group: "Deal Rules",
        },
        // win conditions
        {
          key: "win_set_amount",
          label: "Win set amount",
          min: 3,
          max: 6,
          group: "Win Conditions",
        },
        {
          key: "win_money_amount",
          label: "Win money amount",
          min: 0,
          max: 40,
          group: "Win Conditions",
        },
        // no mercy rules
        {
          key: "debt_chips_per_player",
          label: "Debt chips per player",
          min: 1,
          max: 5,
          group: "No Mercy Rules",
        },
        {
          key: "yoink_payment",
          label: "Yoink! payment",
          min: 5,
          max: 15,
          group: "No Mercy Rules",
        },
        {
          key: "shack_rent_bonus",
          label: "Shack rent bonus",
          min: 3,
          max: 8,
          group: "No Mercy Rules",
        },
        {
          key: "big_payday_hand_target",
          label: "Big Payday hand target",
          min: 5,
          max: 10,
          group: "No Mercy Rules",
        },
        // deck composition
        {
          key: "set_snatcher_amount",
          label: "Set Snatcher cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "debt_trap_amount",
          label: "Debt Trap cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "go_again_amount",
          label: "Go Again! cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "heist_amount",
          label: "Heist cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "market_crash_amount",
          label: "Market Crash cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "big_payday_amount",
          label: "Big Payday cards",
          min: 2,
          max: 6,
          group: "Deck Composition",
        },
        {
          key: "repo_man_amount",
          label: "Repo Man cards",
          min: 1,
          max: 4,
          group: "Deck Composition",
        },
        {
          key: "shack_amount",
          label: "Shack cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "property_raid_amount",
          label: "Property Raid cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "tax_day_amount",
          label: "Tax Day cards",
          min: 1,
          max: 4,
          group: "Deck Composition",
        },
        {
          key: "pickpocket_amount",
          label: "Pickpocket cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "bank_swap_amount",
          label: "Bank Swap cards",
          min: 1,
          max: 4,
          group: "Deck Composition",
        },
        {
          key: "yoink_amount",
          label: "Yoink! cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "nah_amount",
          label: "Nah cards",
          min: 3,
          max: 7,
          group: "Deck Composition",
        },
        {
          key: "rent_amount",
          label: "Rent cards",
          min: 1,
          max: 3,
          group: "Deck Composition",
        },
        {
          key: "double_rent_amount",
          label: "Double Rent cards",
          min: 1,
          max: 3,
          group: "Deck Composition",
        },
        {
          key: "wild_rent_amount",
          label: "Wild Rent cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        {
          key: "double_rent_wild_amount",
          label: "Double Rent Wild cards",
          min: 2,
          max: 5,
          group: "Deck Composition",
        },
        // nah! rules
        {
          key: "nah_consumes_move",
          label: "Nah consumes move",
          group: "Nah! Rules",
          kind: "boolean",
          options: [
            { value: "true", label: "On" },
            { value: "false", label: "Off" },
          ],
        },
      ];
    default:
      return assertNever(game as never);
  }
};

export const getGameSettingDefinitions = (
  gameKey: string,
): GameSettingDefinition[] => {
  const game = parseGame(gameKey);
  return game ? getSettingDefinitionsForGame(game) : [];
};

export const getGameSettingDefinition = (
  game: Game,
  settingKey: string,
): GameSettingDefinition | null => {
  const settings = getSettingDefinitionsForGame(game);
  const matched = settings.find((setting) => setting.key === settingKey);
  return matched ?? null;
};

const parseSettingsObject = (
  settings: EncodedSettings,
): Record<string, unknown> | null => {
  if (settings instanceof Uint8Array) {
    return decodeMsgPackObject(settings);
  }

  if (Array.isArray(settings)) {
    const bytes = toUint8Array(settings);
    if (!bytes) {
      return null;
    }
    return decodeMsgPackObject(bytes);
  }

  if (!settings.trim()) {
    return null;
  }

  try {
    const parsed = JSON.parse(settings) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null;
    }
    return parsed as Record<string, unknown>;
  } catch {
    const bytes = decodeBase64(settings);
    if (!bytes) {
      return null;
    }
    return decodeMsgPackObject(bytes);
  }
};

export const parseGameSettings = <TGame extends Game>(
  game: TGame,
  settings: EncodedSettings,
): GameSettingsFor<TGame> => {
  const defaults = getDefaultSettingsForGame(game);
  const parsed = parseSettingsObject(settings);

  switch (game) {
    case Game.MonopolyDeal: {
      const monopolyDefaults = defaults as MonopolyDealSettings;
      const definitions = getSettingDefinitionsForGame(Game.MonopolyDeal);
      if (!parsed) {
        return defaults;
      }

      const getSettingValue = (key: string, defaultValue: number): number => {
        const definition = definitions.find((setting) => setting.key === key);
        if (!definition || definition.kind === "boolean") {
          return defaultValue;
        }

        const value = parsed[key];
        if (
          typeof value === "number" &&
          inRange(value, definition.min, definition.max)
        ) {
          return value;
        }

        return defaultValue;
      };

      const getBooleanSettingValue = (
        key: string,
        defaultValue: boolean,
      ): boolean => {
        const definition = definitions.find((setting) => setting.key === key);
        if (!definition || definition.kind !== "boolean") {
          return defaultValue;
        }

        const value = parsed[key];
        if (typeof value === "boolean") {
          return value;
        }

        return defaultValue;
      };

      return {
        // deal rules
        num_decks: getSettingValue("num_decks", monopolyDefaults.num_decks),
        start_num_cards: getSettingValue(
          "start_num_cards",
          monopolyDefaults.start_num_cards,
        ),
        max_hand_size: getSettingValue("max_hand_size", monopolyDefaults.max_hand_size),
        moves_per_turn: getSettingValue(
          "moves_per_turn",
          monopolyDefaults.moves_per_turn,
        ),
        // card rules
        payday_draw: getSettingValue("payday_draw", monopolyDefaults.payday_draw),
        party_bill_payment: getSettingValue(
          "party_bill_payment",
          monopolyDefaults.party_bill_payment,
        ),
        settle_up_payment: getSettingValue(
          "settle_up_payment",
          monopolyDefaults.settle_up_payment,
        ),
        // win conditions
        win_set_amount: getSettingValue(
          "win_set_amount",
          monopolyDefaults.win_set_amount,
        ),
        win_money_amount: getSettingValue(
          "win_money_amount",
          monopolyDefaults.win_money_amount,
        ),
        // game speed
        speed: getSettingValue("speed", monopolyDefaults.speed),
        // deck rules
        payday_amount: getSettingValue(
          "payday_amount",
          monopolyDefaults.payday_amount,
        ),
        double_the_rent_amount: getSettingValue(
          "double_the_rent_amount",
          monopolyDefaults.double_the_rent_amount,
        ),
        party_bill_amount: getSettingValue(
          "party_bill_amount",
          monopolyDefaults.party_bill_amount,
        ),
        house_amount: getSettingValue("house_amount", monopolyDefaults.house_amount),
        property_steal_amount: getSettingValue(
          "property_steal_amount",
          monopolyDefaults.property_steal_amount,
        ),
        property_swap_amount: getSettingValue(
          "property_swap_amount",
          monopolyDefaults.property_swap_amount,
        ),
        settle_up_amount: getSettingValue(
          "settle_up_amount",
          monopolyDefaults.settle_up_amount,
        ),
        hotel_amount: getSettingValue("hotel_amount", monopolyDefaults.hotel_amount),
        nah_amount: getSettingValue("nah_amount", monopolyDefaults.nah_amount),
        set_snatcher_amount: getSettingValue(
          "set_snatcher_amount",
          monopolyDefaults.set_snatcher_amount,
        ),
        rent_amount: getSettingValue("rent_amount", monopolyDefaults.rent_amount),
        wild_rent_amount: getSettingValue(
          "wild_rent_amount",
          monopolyDefaults.wild_rent_amount,
        ),
        // nah! rules
        nah_consumes_move: getBooleanSettingValue(
          "nah_consumes_move",
          monopolyDefaults.nah_consumes_move,
        ),
      } as GameSettingsFor<TGame>;
    }
    case Game.DealNoMercy: {
      const definitions = getSettingDefinitionsForGame(Game.DealNoMercy);
      if (!parsed) {
        return defaults;
      }

      // Definition-driven merge: every numeric knob is validated against its
      // range and every boolean knob against its type, falling back to the
      // default otherwise. Keeps this in lockstep with the definitions list
      // above (and, transitively, the engine's validate tags).
      const merged: Record<string, unknown> = {
        ...(defaults as Record<string, unknown>),
      };
      for (const definition of definitions) {
        const value = parsed[definition.key];
        if (definition.kind === "boolean") {
          if (typeof value === "boolean") {
            merged[definition.key] = value;
          }
          continue;
        }
        if (
          typeof value === "number" &&
          inRange(value, definition.min, definition.max)
        ) {
          merged[definition.key] = value;
        }
      }

      return merged as GameSettingsFor<TGame>;
    }
    default:
      return assertNever(game as never);
  }
};

export const getGameSettingSelectValues = (
  gameKey: string,
  settings: EncodedSettings,
): GameSettingSelectValue[] => {
  const game = parseGame(gameKey);
  if (!game) {
    return [];
  }

  switch (game) {
    case Game.MonopolyDeal:
    case Game.DealNoMercy: {
      const parsed = parseGameSettings(game, settings) as Record<
        string,
        unknown
      >;
      const definitions = getSettingDefinitionsForGame(game);

      return definitions.map((definition) => {
        const value = parsed[definition.key];
        if (definition.kind === "boolean") {
          return {
            key: definition.key,
            label: definition.label,
            group: definition.group,
            kind: definition.kind,
            value: String(typeof value === "boolean" ? value : true),
            options: definition.options,
          };
        }

        const numericValue = typeof value === "number" ? value : definition.min;

        return {
          key: definition.key,
          label: definition.label,
          group: definition.group,
          kind: "number" as const,
          value: String(numericValue),
          options:
            definition.key === "speed"
              ? buildSpeedOptions()
              : buildRangeOptions(definition.min, definition.max),
        };
      });
    }
    default:
      return assertNever(game as never);
  }
};

export const getCapacityOptions = (
  min: number,
  max: number,
): Array<{ value: string; label: string }> => {
  return buildRangeOptions(min, max);
};

export const getCapacityRangeForGame = (
  gameKey: string,
  settings: EncodedSettings,
  baseRange: CapacityRange = { min: 2, max: 5 },
): CapacityRange => {
  const game = parseGame(gameKey);
  if (!game) {
    return baseRange;
  }

  switch (game) {
    case Game.MonopolyDeal:
    case Game.DealNoMercy: {
      const parsed = parseGameSettings(game, settings);
      return {
        min: baseRange.min,
        max: baseRange.max * parsed.num_decks,
      };
    }
    default:
      return assertNever(game as never);
  }
};
