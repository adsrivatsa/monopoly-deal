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
      return { num_decks: 1 } as GameSettingsFor<TGame>;
    default:
      return assertNever(game as never);
  }
};

export const stringifyGameSettings = <TGame extends Game>(
  _game: TGame,
  settings: GameSettingsFor<TGame>,
): string => {
  return JSON.stringify(settings);
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
      ];
    default:
      return assertNever(game as never);
  }
};

export const getGameSettingDefinitions = (gameKey: string): GameSettingDefinition[] => {
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

const parseSettingsObject = (settings: string): Record<string, unknown> | null => {
  try {
    const parsed = JSON.parse(settings) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null;
    }
    return parsed as Record<string, unknown>;
  } catch {
    return null;
  }
};

export const parseGameSettings = <TGame extends Game>(
  game: TGame,
  settings: string,
): GameSettingsFor<TGame> => {
  const defaults = getDefaultSettingsForGame(game);
  const parsed = parseSettingsObject(settings);

  switch (game) {
    case Game.MonopolyDeal: {
      const numDecksSetting = getSettingDefinitionsForGame(Game.MonopolyDeal)[0];
      if (!parsed) {
        return defaults;
      }

      const numDecks = parsed.num_decks;
      if (
        typeof numDecks === "number" &&
        inRange(numDecks, numDecksSetting.min, numDecksSetting.max)
      ) {
        return { num_decks: numDecks } as GameSettingsFor<TGame>;
      }

      return defaults;
    }
    default:
      return assertNever(game as never);
  }
};

export const getGameSettingSelectValues = (
  gameKey: string,
  settings: string,
): GameSettingSelectValue[] => {
  const game = parseGame(gameKey);
  if (!game) {
    return [];
  }

  switch (game) {
    case Game.MonopolyDeal: {
      const parsed = parseGameSettings(game, settings);
      const definition = getSettingDefinitionsForGame(game)[0];
      return [
        {
          key: definition.key,
          label: definition.label,
          value: String(parsed.num_decks),
          options: buildRangeOptions(definition.min, definition.max),
        },
      ];
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
  settings: string,
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
