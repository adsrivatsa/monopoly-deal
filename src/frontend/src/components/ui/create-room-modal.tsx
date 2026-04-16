import { FormEvent, MouseEvent, useEffect, useMemo, useState } from "react";
import type { CreateRoomParams } from "../../api/room";
import {
  getCapacityOptions,
  getCapacityRangeForGame,
  Game,
  getDefaultSettingsForGame,
  getGameDisplayName,
  getGameSettingDefinition,
  parseGame,
  stringifyGameSettings,
} from "../../api/models";
import { appConfig } from "../../config";
import Button from "./button";

type CreateRoomModalProps = {
  onClose: () => void;
  onCreate: (values: CreateRoomParams) => void | Promise<void>;
};

const configuredGames: Game[] = appConfig.room.create.games
  .map((game) => parseGame(game))
  .filter((game): game is Game => game !== null);

const availableGames: Game[] =
  configuredGames.length > 0 ? configuredGames : [Game.MonopolyDeal];

const gameOptions: Array<{ value: Game; label: string }> = availableGames.map((game) => {
  return {
    value: game,
    label: getGameDisplayName(game),
  };
});

const monopolyDeckSetting = getGameSettingDefinition(
  Game.MonopolyDeal,
  "num_decks",
);

const CreateRoomModal = ({ onClose, onCreate }: CreateRoomModalProps) => {
  const [displayName, setDisplayName] = useState("");
  const [capacity, setCapacity] = useState(String(appConfig.room.create.capacity.min));
  const [game, setGame] = useState<Game>(availableGames[0]);
  const [monopolyNumDecks, setMonopolyNumDecks] = useState(
    String(getDefaultSettingsForGame(Game.MonopolyDeal).num_decks),
  );

  const currentSettings = useMemo(() => {
    if (game === Game.MonopolyDeal) {
      const parsedNumDecks = Number.parseInt(monopolyNumDecks, 10);
      return stringifyGameSettings(game, {
        num_decks: Number.isNaN(parsedNumDecks) ? 1 : parsedNumDecks,
      });
    }

    return stringifyGameSettings(game, getDefaultSettingsForGame(game));
  }, [game, monopolyNumDecks]);

  const capacityRange = useMemo(() => {
    return getCapacityRangeForGame(game, currentSettings, {
      min: appConfig.room.create.capacity.min,
      max: appConfig.room.create.capacity.max,
    });
  }, [currentSettings, game]);

  useEffect(() => {
    const parsedCapacity = Number.parseInt(capacity, 10);
    if (
      Number.isNaN(parsedCapacity) ||
      parsedCapacity < capacityRange.min ||
      parsedCapacity > capacityRange.max
    ) {
      setCapacity(String(capacityRange.min));
    }
  }, [capacity, capacityRange.max, capacityRange.min]);

  const onCardClick = (event: MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();
  };

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const selectedGame = parseGame(game);
    const parsedCapacity = Number.parseInt(capacity, 10);
    if (
      !selectedGame ||
      !displayName.trim() ||
      Number.isNaN(parsedCapacity) ||
      parsedCapacity < capacityRange.min ||
      parsedCapacity > capacityRange.max
    ) {
      return;
    }

    const defaultSettings = getDefaultSettingsForGame(selectedGame);
    let settings = stringifyGameSettings(selectedGame, defaultSettings);

    if (selectedGame === Game.MonopolyDeal) {
      const parsedNumDecks = Number.parseInt(monopolyNumDecks, 10);
      const minDecks = monopolyDeckSetting?.min ?? 1;
      const maxDecks = monopolyDeckSetting?.max ?? 5;
      if (
        Number.isNaN(parsedNumDecks) ||
        parsedNumDecks < minDecks ||
        parsedNumDecks > maxDecks
      ) {
        return;
      }

      settings = stringifyGameSettings(selectedGame, {
        num_decks: parsedNumDecks,
      });
    }

    onCreate({
      display_name: displayName.trim(),
      capacity: parsedCapacity,
      game: selectedGame,
      settings,
    });
  };

  return (
    <div className="overlay-backdrop" onClick={onClose} role="presentation">
      <div
        className="overlay-card"
        onClick={onCardClick}
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-room-title"
      >
        <p className="eyebrow">Lobby</p>
        <h2 id="create-room-title" className="overlay-title">
          Create room
        </h2>

        <form onSubmit={onSubmit} className="form-stack">
          <label className="field-label" htmlFor="room-display-name">
            Display name
          </label>
          <input
            id="room-display-name"
            className="field-input"
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder="Enter room name"
            autoFocus
          />

          <label className="field-label" htmlFor="room-capacity">
            Capacity
          </label>
          <select
            id="room-capacity"
            className="field-input"
            value={capacity}
            onChange={(event) => setCapacity(event.target.value)}
          >
            {getCapacityOptions(capacityRange.min, capacityRange.max).map(
              (option) => {
                return (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                );
              },
            )}
          </select>

          <label className="field-label" htmlFor="room-game">
            Game
          </label>
          <select
            id="room-game"
            className="field-input"
            value={game}
            onChange={(event) => {
              const nextGame = parseGame(event.target.value);
              if (!nextGame) {
                return;
              }
              setGame(nextGame);
            }}
          >
            {gameOptions.map((option) => {
              return (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              );
            })}
          </select>

          {game === Game.MonopolyDeal ? (
            <>
              <label className="field-label" htmlFor="monopoly-num-decks">
                Number of decks
              </label>
              <select
                id="monopoly-num-decks"
                className="field-input"
                aria-label="Select number of decks"
                value={monopolyNumDecks}
                onChange={(event) => setMonopolyNumDecks(event.target.value)}
              >
                {getCapacityOptions(
                  monopolyDeckSetting?.min ?? 1,
                  monopolyDeckSetting?.max ?? 5,
                ).map((option) => {
                  return (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  );
                })}
              </select>
            </>
          ) : null}

          <div className="overlay-actions">
            <Button variant="outline" type="button" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">Create room</Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateRoomModal;
