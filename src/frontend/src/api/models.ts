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
  num_decks: number;
  start_num_cards: number;
  max_hand_size: number;
  moves_per_turn: number;
  pass_go_draw: number;
  its_my_birthday_amount: number;
  debt_collector_amount: number;
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
};

export type GameSettingSelectValue = {
  key: string;
  label: string;
  value: string;
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
        num_decks: 1,
        start_num_cards: 5,
        max_hand_size: 7,
        moves_per_turn: 3,
        pass_go_draw: 2,
        its_my_birthday_amount: 2,
        debt_collector_amount: 5,
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
        {
          key: "num_decks",
          label: "Number of decks",
          min: 1,
          max: 3,
        },
        {
          key: "start_num_cards",
          label: "Starting cards",
          min: 5,
          max: 8,
        },
        {
          key: "max_hand_size",
          label: "Max hand size",
          min: 7,
          max: 10,
        },
        {
          key: "moves_per_turn",
          label: "Moves per turn",
          min: 3,
          max: 5,
        },
        {
          key: "pass_go_draw",
          label: "Pass Go draw",
          min: 2,
          max: 5,
        },
        {
          key: "its_my_birthday_amount",
          label: "It's My Birthday amount",
          min: 2,
          max: 5,
        },
        {
          key: "debt_collector_amount",
          label: "Debt Collector amount",
          min: 5,
          max: 8,
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
        pass_go_draw: getSettingValue("pass_go_draw", defaults.pass_go_draw),
        its_my_birthday_amount: getSettingValue(
          "its_my_birthday_amount",
          defaults.its_my_birthday_amount,
        ),
        debt_collector_amount: getSettingValue(
          "debt_collector_amount",
          defaults.debt_collector_amount,
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
          value: String(numericValue),
          options: buildRangeOptions(definition.min, definition.max),
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
