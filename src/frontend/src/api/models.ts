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
}

export const supportedGames: Game[] = [Game.MonopolyDeal];

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
};

type GameSettingsByGame = {
  [Game.MonopolyDeal]: MonopolyDealSettings;
};

export type GameSettingsFor<TGame extends Game> = GameSettingsByGame[TGame];

export type GameSettingDefinition = {
  key: string;
  label: string;
  min: number;
  max: number;
  group: string;
};

export type GameSettingSelectValue = {
  key: string;
  label: string;
  value: string;
  group: string;
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
};

export const parseGame = (gameKey: string): Game | null => {
  switch (gameKey) {
    case Game.MonopolyDeal:
      return Game.MonopolyDeal;
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
      const definitions = getSettingDefinitionsForGame(Game.MonopolyDeal);
      if (!parsed) {
        return defaults;
      }

      const getSettingValue = (key: string, defaultValue: number): number => {
        const definition = definitions.find((setting) => setting.key === key);
        if (!definition) {
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

      return {
        // deal rules
        num_decks: getSettingValue("num_decks", defaults.num_decks),
        start_num_cards: getSettingValue(
          "start_num_cards",
          defaults.start_num_cards,
        ),
        max_hand_size: getSettingValue("max_hand_size", defaults.max_hand_size),
        moves_per_turn: getSettingValue(
          "moves_per_turn",
          defaults.moves_per_turn,
        ),
        // card rules
        payday_draw: getSettingValue("payday_draw", defaults.payday_draw),
        party_bill_payment: getSettingValue(
          "party_bill_payment",
          defaults.party_bill_payment,
        ),
        settle_up_payment: getSettingValue(
          "settle_up_payment",
          defaults.settle_up_payment,
        ),
        // win conditions
        win_set_amount: getSettingValue(
          "win_set_amount",
          defaults.win_set_amount,
        ),
        win_money_amount: getSettingValue(
          "win_money_amount",
          defaults.win_money_amount,
        ),
        // game speed
        speed: getSettingValue("speed", defaults.speed),
        // deck rules
        payday_amount: getSettingValue(
          "payday_amount",
          defaults.payday_amount,
        ),
        double_the_rent_amount: getSettingValue(
          "double_the_rent_amount",
          defaults.double_the_rent_amount,
        ),
        party_bill_amount: getSettingValue(
          "party_bill_amount",
          defaults.party_bill_amount,
        ),
        house_amount: getSettingValue("house_amount", defaults.house_amount),
        property_steal_amount: getSettingValue(
          "property_steal_amount",
          defaults.property_steal_amount,
        ),
        property_swap_amount: getSettingValue(
          "property_swap_amount",
          defaults.property_swap_amount,
        ),
        settle_up_amount: getSettingValue(
          "settle_up_amount",
          defaults.settle_up_amount,
        ),
        hotel_amount: getSettingValue("hotel_amount", defaults.hotel_amount),
        nah_amount: getSettingValue("nah_amount", defaults.nah_amount),
        set_snatcher_amount: getSettingValue(
          "set_snatcher_amount",
          defaults.set_snatcher_amount,
        ),
        rent_amount: getSettingValue("rent_amount", defaults.rent_amount),
        wild_rent_amount: getSettingValue(
          "wild_rent_amount",
          defaults.wild_rent_amount,
        ),
      } as GameSettingsFor<TGame>;
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
    case Game.MonopolyDeal: {
      const parsed = parseGameSettings(game, settings);
      const definitions = getSettingDefinitionsForGame(game);

      return definitions.map((definition) => {
        const value = parsed[definition.key as keyof MonopolyDealSettings];
        const numericValue = typeof value === "number" ? value : definition.min;

        return {
          key: definition.key,
          label: definition.label,
          group: definition.group,
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
    case Game.MonopolyDeal: {
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
