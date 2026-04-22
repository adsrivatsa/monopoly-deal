import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  connectGameSocket,
  decodeGameServerMessage,
  sendGameChatMessage,
  sendGameComplyPaymentDemandMessage,
  sendGameCompleteTurnMessage,
  sendGamePlayMoneyMessage,
  sendGamePlayItsMyBirthdayMessage,
  sendGamePlayDebtCollectorMessage,
  sendGamePlayDoubleTheRentMessage,
  sendGamePlayPassGoMessage,
  sendGamePlayRentMessage,
  sendGamePlaySlyDealMessage,
  sendGamePlayForcedDealMessage,
  sendGamePlayDealBreakerMessage,
  sendGamePlayHouseMessage,
  sendGamePlayHotelMessage,
  sendGamePlayWildRentMessage,
  sendGameComplyPropertyDemandMessage,
  sendGameComplyPropertySetDemandMessage,
  sendGameRearrangeCardMessage,
  sendGameDiscardCardsMessage,
  sendGameDenyDemandMessage,
  sendGameResolvePendingRentMessage,
  sendGamePlayPropertyMessage,
  toGameServerMessageJson,
} from "../api/gameSocket";
import { getPlayer } from "../api/player";
import ChatBox from "../components/chat/ChatBox";
import TurnControlsCard from "../components/game/TurnControlsCard";
import Button from "../components/ui/button";
import ErrorToastStack, { type ErrorToastNotice } from "../components/ui/error-toast-stack";
import MonopolyDealGameMount from "../game/monopoly_deal/MonopolyDealGameMount";
import {
  AssetKey,
  Category,
  Color,
  type Error as GameError,
  type AssetImage,
  type Card,
  type GameState,
  type Money,
  type PropertySet,
  type Player,
  type TransferCards,
  type WonGame,
} from "../generated/monopoly_deal";

type GameChatMessage = {
  id: string;
  playerId: string;
  text: string;
};

type GameErrorNotice = ErrorToastNotice;

type WonGameResult = WonGame & {
  displayName?: string;
  avatarUrl?: string;
};

const toAssetImageMap = (assetImages: AssetImage[]): Record<number, string> => {
  return assetImages.reduce<Record<number, string>>((lookup, assetImage) => {
    if (assetImage.imageUrl) {
      lookup[assetImage.assetKey] = assetImage.imageUrl;
    }

    return lookup;
  }, {});
};

const toClientGameError = (error: unknown, code: string): GameError => {
  if (error && typeof error === "object" && "message" in error) {
    const message = String((error as { message: unknown }).message);
    return {
      message,
      code,
      status: 0,
    };
  }

  return {
    message: "Unexpected client-side error.",
    code,
    status: 0,
  };
};

const isGameErrorLike = (value: unknown): value is GameError => {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<GameError>;
  return typeof candidate.message === "string";
};

const toServerGameErrors = (value: unknown): GameError[] => {
  const collected: GameError[] = [];
  const visited = new Set<unknown>();

  const visit = (node: unknown) => {
    if (!node || typeof node !== "object" || visited.has(node)) {
      return;
    }
    visited.add(node);

    if (isGameErrorLike(node)) {
      collected.push({
        message: node.message,
        code: typeof node.code === "string" ? node.code : "SERVER_ERROR",
        status: typeof node.status === "number" ? node.status : 0,
      });
    }

    if (Array.isArray(node)) {
      for (const item of node) {
        visit(item);
      }
      return;
    }

    for (const [key, child] of Object.entries(node)) {
      if (key.toLowerCase().includes("error")) {
        visit(child);
        continue;
      }

      if (typeof child === "object") {
        visit(child);
      }
    }
  };

  visit(value);

  return collected.filter((error) => error.message.trim().length > 0);
};

const toCardIdSet = (cards: Card[]): Set<string> => {
  return new Set(cards.map((card) => card.cardId));
};

const toPropertySetCardIdSet = (propertySets: PropertySet[]): Set<string> => {
  const cardIds = new Set<string>();
  for (const propertySet of propertySets) {
    for (const card of propertySet.cards) {
      cardIds.add(card.cardId);
    }
  }
  return cardIds;
};

const isNormalRentAssetKey = (assetKey: AssetKey): boolean => {
  return (
    assetKey === AssetKey.ASSET_KEY_RENT_BROWN_SKY ||
    assetKey === AssetKey.ASSET_KEY_RENT_PINK_ORANGE ||
    assetKey === AssetKey.ASSET_KEY_RENT_RED_YELLOW ||
    assetKey === AssetKey.ASSET_KEY_RENT_GREEN_BLUE ||
    assetKey === AssetKey.ASSET_KEY_RENT_UTILITY_RAILROAD
  );
};

const minPropertyCountForCompleteSet = (color: Color): number => {
  switch (color) {
    case Color.COLOR_BROWN:
    case Color.COLOR_BLUE:
    case Color.COLOR_UTILITY:
      return 2;
    case Color.COLOR_SKY:
    case Color.COLOR_PINK:
    case Color.COLOR_ORANGE:
    case Color.COLOR_RED:
    case Color.COLOR_YELLOW:
    case Color.COLOR_GREEN:
      return 3;
    case Color.COLOR_RAILROAD:
      return 4;
    default:
      return Number.POSITIVE_INFINITY;
  }
};

const isPropertyCardForSetCompletion = (card: Card): boolean => {
  return (
    card.category === Category.CATEGORY_PURE_PROPERTY ||
    card.category === Category.CATEGORY_WILD_PROPERTY
  );
};

const isPropertySetCompleteForBuilding = (propertySet: PropertySet): boolean => {
  const requiredCount = minPropertyCountForCompleteSet(propertySet.color);
  if (!Number.isFinite(requiredCount)) {
    return false;
  }

  const propertyCardCount = propertySet.cards.filter((card) => {
    return isPropertyCardForSetCompletion(card);
  }).length;

  return propertyCardCount >= requiredCount;
};

const sumCardValues = (cards: Card[]): number => {
  return cards.reduce((total, card) => total + (card.value ?? 0), 0);
};

const recalculatePlayerStatsFromBoard = (
  players: Player[],
  moneyPiles: Money[],
  propertySets: PropertySet[],
): Player[] => {
  return players.map((player) => {
    const moneyValue = moneyPiles
      .filter((pile) => pile.playerId === player.playerId)
      .reduce((total, pile) => total + sumCardValues(pile.cards), 0);
    const propertyValue = propertySets
      .filter((set) => set.playerId === player.playerId)
      .reduce((total, set) => total + sumCardValues(set.cards), 0);
    const completedSets = propertySets.filter((set) => {
      return (
        set.playerId === player.playerId && isPropertySetCompleteForBuilding(set)
      );
    }).length;

    return {
      ...player,
      money: moneyValue + propertyValue,
      completedSets,
    };
  });
};

const getTransferCardsPlayerStats = (transferCards: TransferCards) => {
  const stats = transferCards as TransferCards & {
    sourceProperties?: number;
    targetProperties?: number;
  };

  const sourceCompletedSets =
    typeof stats.sourceProperties === "number"
      ? stats.sourceProperties
      : stats.sourceSets;
  const targetCompletedSets =
    typeof stats.targetProperties === "number"
      ? stats.targetProperties
      : stats.targetSets;

  return {
    sourceCompletedSets,
    targetCompletedSets,
    sourceMoney: stats.sourceMoney,
    targetMoney: stats.targetMoney,
  };
};

const applyTransferCardsPlayerStats = (
  players: Player[],
  sourceId: string,
  targetId: string,
  sourceCompletedSets: number,
  targetCompletedSets: number,
  sourceMoney: number,
  targetMoney: number,
): Player[] => {
  return players.map((player) => {
    if (player.playerId === sourceId) {
      return {
        ...player,
        completedSets: sourceCompletedSets,
        money: sourceMoney,
      };
    }

    if (player.playerId === targetId) {
      return {
        ...player,
        completedSets: targetCompletedSets,
        money: targetMoney,
      };
    }

    return player;
  });
};

let globalSocketSessionId = 0;
let globalGameSocket: WebSocket | null = null;
let globalReconnectTimeoutId: number | null = null;

const GamePage = () => {
  const { game_id: gameId } = useParams();
  const navigate = useNavigate();
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectAttemptRef = useRef(0);
  const [initialGameState, setInitialGameState] = useState<GameState | null>(
    null,
  );
  const [assetImageByKey, setAssetImageByKey] = useState<
    Record<number, string>
  >({});
  const [selfPlayerId, setSelfPlayerId] = useState<string | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);
  const [currentTurnPlayerId, setCurrentTurnPlayerId] = useState<string | null>(
    null,
  );
  const [movesLeft, setMovesLeft] = useState(0);
  const [playerNameById, setPlayerNameById] = useState<Record<string, string>>(
    {},
  );
  const [chatMessages, setChatMessages] = useState<GameChatMessage[]>([]);
  const [errorNotices, setErrorNotices] = useState<GameErrorNotice[]>([]);
  const [selectedDiscardCardIds, setSelectedDiscardCardIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [wonGame, setWonGame] = useState<WonGameResult | null>(null);

  const pushErrorNotice = useCallback(
    (
      error: Pick<GameError, "message" | "code">,
      options?: {
        title?: string;
        eyebrow?: string;
      },
    ) => {
      setErrorNotices((current) => {
        const nextNotice: GameErrorNotice = {
          id: `error-${Date.now()}-${Math.random().toString(16).slice(2)}`,
          message: error.message,
          code: error.code,
          title: options?.title,
          eyebrow: options?.eyebrow,
        };

        return [...current.slice(-4), nextNotice];
      });
    },
    [],
  );

  const dismissErrorNotice = useCallback((id: string) => {
    setErrorNotices((current) => current.filter((notice) => notice.id !== id));
  }, []);

  const notifyFrontendRule = useCallback(
    (message: string, code: string) => {
      pushErrorNotice(
        { message, code },
        { title: "Action blocked", eyebrow: "Frontend rule" },
      );
    },
    [pushErrorNotice],
  );

  const notifySocketUnavailable = useCallback(() => {
    pushErrorNotice(
      {
        message: "Game connection is not available right now.",
        code: "SOCKET_NOT_OPEN",
      },
      { title: "Connection issue", eyebrow: "Network" },
    );
  }, [pushErrorNotice]);

  const handCards = initialGameState?.yourHand?.cards ?? [];
  const handCardIdsKey = handCards.map((card) => card.cardId).join("|");
  const maxHandSize = initialGameState?.maxHandSize ?? 0;
  const discardRequiredCount = Math.max(0, handCards.length - maxHandSize);
  const isSelfTurn = !!selfPlayerId && selfPlayerId === currentTurnPlayerId;
  const isDiscardRequired =
    isSelfTurn &&
    movesLeft <= 0 &&
    discardRequiredCount > 0;

  useEffect(() => {
    void (async () => {
      try {
        const response = await getPlayer();
        if (!response.ok) {
          pushErrorNotice({
            message: "Could not load current player details.",
            code: "PLAYER_LOOKUP_FAILED",
          }, {
            title: "Game setup failed",
            eyebrow: "Player error",
          });
          return;
        }

        const data = (await response.json()) as { player_id: string };
        setSelfPlayerId(data.player_id);
      } catch (error) {
        pushErrorNotice(toClientGameError(error, "PLAYER_LOOKUP_CRASH"), {
          title: "Game setup failed",
          eyebrow: "Player error",
        });
      }
    })();
  }, [pushErrorNotice]);

  useEffect(() => {
    const nav = document.querySelector<HTMLElement>(".app-nav");
    if (!nav) {
      return;
    }

    const rootStyle = document.documentElement.style;
    const updateOffset = () => {
      const navHeight = Math.ceil(nav.getBoundingClientRect().height);
      rootStyle.setProperty("--game-nav-offset", `${navHeight}px`);
    };

    updateOffset();
    const resizeObserver = new ResizeObserver(updateOffset);
    resizeObserver.observe(nav);
    window.addEventListener("resize", updateOffset);

    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", updateOffset);
      rootStyle.removeProperty("--game-nav-offset");
    };
  }, []);

  useEffect(() => {
    let didUnmount = false;
    let activeSocket: WebSocket | null = null;
    const sessionId = globalSocketSessionId + 1;
    globalSocketSessionId = sessionId;

    const isCurrentSession = () => {
      return !didUnmount && globalSocketSessionId === sessionId;
    };

    const clearReconnectTimer = () => {
      if (globalReconnectTimeoutId !== null) {
        window.clearTimeout(globalReconnectTimeoutId);
        globalReconnectTimeoutId = null;
      }
    };

    const scheduleReconnect = () => {
      if (!isCurrentSession()) {
        return;
      }

      clearReconnectTimer();
      const attempt = reconnectAttemptRef.current;
      const delayMs = Math.min(5000, 500 * 2 ** attempt);
      reconnectAttemptRef.current = attempt + 1;

      console.log("[game-ws] reconnect scheduled", {
        attempt: attempt + 1,
        delayMs,
      });

      globalReconnectTimeoutId = window.setTimeout(() => {
        globalReconnectTimeoutId = null;
        if (!isCurrentSession()) {
          return;
        }
        connectSocket();
      }, delayMs);
    };

    const connectSocket = () => {
      if (!isCurrentSession()) {
        return;
      }

      const existingSocket = globalGameSocket;
      if (existingSocket && existingSocket !== activeSocket) {
        socketRef.current = null;
        existingSocket.close();
      }

      const socket = connectGameSocket();
      activeSocket = socket;
      globalGameSocket = socket;
      socketRef.current = socket;

      console.log("[game-ws] connecting", socket.url);

      socket.onopen = () => {
        if (
          !isCurrentSession() ||
          socketRef.current !== socket ||
          globalGameSocket !== socket
        ) {
          return;
        }
        reconnectAttemptRef.current = 0;
        clearReconnectTimer();
        console.log("[game-ws] open", { gameId });
      };

      socket.onmessage = (event) => {
        if (
          !isCurrentSession() ||
          socketRef.current !== socket ||
          globalGameSocket !== socket
        ) {
          return;
        }

        void (async () => {
          try {
          const message = await decodeGameServerMessage(event.data);
          if (!message) {
            console.log("[game-ws] message (non-binary)", event.data);
            if (typeof event.data === "string") {
              try {
                const parsed = JSON.parse(event.data) as unknown;
                const serverErrors = toServerGameErrors(parsed);
                for (const serverError of serverErrors) {
                  pushErrorNotice(serverError, {
                    title: "Server error",
                    eyebrow: "Game server",
                  });
                }
              } catch {
                pushErrorNotice({
                  message: event.data,
                  code: "SERVER_TEXT_EVENT",
                }, {
                  title: "Server error",
                  eyebrow: "Game server",
                });
              }
            }
            return;
          }

          const serverErrors = toServerGameErrors(toGameServerMessageJson(message));
          for (const serverError of serverErrors) {
            pushErrorNotice(serverError, {
              title: "Server error",
              eyebrow: "Game server",
            });
          }

          const assetImages =
            message.monopolyDealMessage?.gameState?.assetImages;
          if (assetImages && assetImages.length > 0) {
            const incomingAssetImageByKey = toAssetImageMap(assetImages);
            setAssetImageByKey((current) => {
              let hasChanges = false;

              for (const [assetKey, imageUrl] of Object.entries(
                incomingAssetImageByKey,
              )) {
                if (current[Number(assetKey)] !== imageUrl) {
                  hasChanges = true;
                  break;
                }
              }

              if (!hasChanges) {
                return current;
              }

              return {
                ...current,
                ...incomingAssetImageByKey,
              };
            });
          }

          const gameState = message.monopolyDealMessage?.gameState;
          if (gameState) {
            setPlayers(gameState.players);
            setCurrentTurnPlayerId(gameState.currentPlayerId);
            setMovesLeft(gameState.movesLeft);
            setPlayerNameById((current) => {
              const incoming = Object.fromEntries(
                gameState.players.map((player) => [
                  player.playerId,
                  player.displayName,
                ]),
              );

              return {
                ...current,
                ...incoming,
              };
            });

            setInitialGameState((current) => {
              if (current) {
                return current;
              }

              console.log("[game-ws] initial game state", gameState);
              return gameState;
            });
          }

          const chatReceived = message.monopolyDealMessage?.chatReceived;
          if (chatReceived) {
            setChatMessages((current) => {
              return [
                ...current,
                {
                  id: `chat-${chatReceived.playerId}-${Date.now()}-${current.length}`,
                  playerId: chatReceived.playerId,
                  text: chatReceived.payload,
                },
              ];
            });
          }

          const wonGameMessage = message.monopolyDealMessage?.wonGame as
            | WonGameResult
            | undefined;
          if (wonGameMessage) {
            setWonGame(wonGameMessage);
          }

          const startTurnRes = message.monopolyDealMessage?.startTurnRes;
          if (startTurnRes) {
            setCurrentTurnPlayerId(startTurnRes.playerId);
            setMovesLeft(startTurnRes.movesLeft);

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const isSelfStartTurn = selfPlayerId === startTurnRes.playerId;
              const drawnCards = startTurnRes.cards ?? [];
              const nextYourHand = isSelfStartTurn
                ? {
                    ...current.yourHand,
                    cards: [...(current.yourHand?.cards ?? []), ...drawnCards],
                  }
                : current.yourHand;
              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== startTurnRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: player.handCards + drawnCards.length,
                };
              });

              return {
                ...current,
                seqNum: startTurnRes.seqNum,
                currentPlayerId: startTurnRes.playerId,
                movesLeft: startTurnRes.movesLeft,
                players: nextPlayers,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== startTurnRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: player.handCards + startTurnRes.cards.length,
                };
              });
            });
          }

          const startTurnMaskedRes =
            message.monopolyDealMessage?.startTurnMaskedRes;
          if (startTurnMaskedRes) {
            setCurrentTurnPlayerId(startTurnMaskedRes.playerId);

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== startTurnMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: player.handCards + startTurnMaskedRes.numCards,
                };
              });

              return {
                ...current,
                seqNum: startTurnMaskedRes.seqNum,
                currentPlayerId: startTurnMaskedRes.playerId,
                players: nextPlayers,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== startTurnMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: player.handCards + startTurnMaskedRes.numCards,
                };
              });
            });
          }

          const playMoneyRes = message.monopolyDealMessage?.playMoneyRes;
          if (
            playMoneyRes &&
            selfPlayerId &&
            playMoneyRes.playerId === selfPlayerId
          ) {
            setMovesLeft((current) => Math.max(0, current - 1));
          }

          if (playMoneyRes?.card) {
            const playedMoneyCard = playMoneyRes.card;
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextMoney = [...current.money];
              const moneyIndex = nextMoney.findIndex(
                (pile) => pile.playerId === playMoneyRes.playerId,
              );

              if (moneyIndex === -1) {
                nextMoney.push({
                  playerId: playMoneyRes.playerId,
                  cards: [playedMoneyCard],
                });
              } else {
                const existingPile = nextMoney[moneyIndex];
                nextMoney[moneyIndex] = {
                  ...existingPile,
                  cards: [...existingPile.cards, playedMoneyCard],
                };
              }

              const isSelfPlay = selfPlayerId === playMoneyRes.playerId;
              const nextYourHand = isSelfPlay
                ? {
                    ...current.yourHand,
                    cards:
                      current.yourHand?.cards.filter(
                        (card) => card.cardId !== playedMoneyCard.cardId,
                      ) ?? [],
                  }
                : current.yourHand;

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== playMoneyRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  money: player.money + (playedMoneyCard.value ?? 0),
                  handCards: Math.max(0, player.handCards - 1),
                };
              });

              return {
                ...current,
                seqNum: playMoneyRes.seqNum,
                players: nextPlayers,
                money: nextMoney,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== playMoneyRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  money: player.money + (playedMoneyCard.value ?? 0),
                  handCards: Math.max(0, player.handCards - 1),
                };
              });
            });
          }

          const playPropertyRes = message.monopolyDealMessage?.playPropertyRes;
          if (
            playPropertyRes &&
            selfPlayerId &&
            playPropertyRes.playerId === selfPlayerId
          ) {
            setMovesLeft((current) => Math.max(0, current - 1));
          }

          if (playPropertyRes?.propertySet) {
            const playedPropertySet = playPropertyRes.propertySet;
            let nextCompletedSetsForPlayer: number | null = null;
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const playedCardId =
                playedPropertySet.cards.at(-1)?.cardId;

              const nextProperties = [...current.properties];
              const propertyIndex = nextProperties.findIndex((propertySet) => {
                return (
                  propertySet.propertySetId ===
                  playedPropertySet.propertySetId
                );
              });

              if (propertyIndex === -1) {
                nextProperties.push(playedPropertySet);
              } else {
                nextProperties[propertyIndex] = playedPropertySet;
              }

              const isSelfPlay = selfPlayerId === playPropertyRes.playerId;
              const nextYourHand =
                isSelfPlay && playedCardId
                  ? {
                      ...current.yourHand,
                      cards:
                        current.yourHand?.cards.filter(
                          (card) => card.cardId !== playedCardId,
                        ) ?? [],
                    }
                  : current.yourHand;

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== playPropertyRes.playerId) {
                  return player;
                }

                const completedSets = nextProperties.filter((propertySet) => {
                  return (
                    propertySet.playerId === player.playerId &&
                    isPropertySetCompleteForBuilding(propertySet)
                  );
                }).length;
                nextCompletedSetsForPlayer = completedSets;

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                  completedSets,
                };
              });

              return {
                ...current,
                seqNum: playPropertyRes.seqNum,
                players: nextPlayers,
                properties: nextProperties,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== playPropertyRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                  completedSets:
                    nextCompletedSetsForPlayer ?? player.completedSets,
                };
              });
            });
          }

          const housePlayed = message.monopolyDealMessage?.housePlayed;
          if (housePlayed) {
            if (selfPlayerId && housePlayed.playerId === selfPlayerId) {
              setMovesLeft((current) => Math.max(0, current - 1));
            }

            setInitialGameState((current) => {
              if (!current || !housePlayed.propertySet) {
                return current;
              }

              const nextProperties = [...current.properties];
              const propertyIndex = nextProperties.findIndex((propertySet) => {
                return propertySet.propertySetId === housePlayed.propertySet?.propertySetId;
              });

              if (propertyIndex === -1) {
                nextProperties.push(housePlayed.propertySet);
              } else {
                nextProperties[propertyIndex] = housePlayed.propertySet;
              }

              const playedCardId = housePlayed.card?.cardId;
              const nextYourHand =
                selfPlayerId === housePlayed.playerId && playedCardId
                  ? {
                      ...current.yourHand,
                      cards:
                        current.yourHand?.cards.filter((card) => {
                          return card.cardId !== playedCardId;
                        }) ?? [],
                    }
                  : current.yourHand;

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== housePlayed.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                };
              });

              return {
                ...current,
                seqNum: housePlayed.seqNum,
                players: nextPlayers,
                properties: nextProperties,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== housePlayed.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                };
              });
            });
          }

          const hotelPlayed = message.monopolyDealMessage?.hotelPlayed;
          if (hotelPlayed) {
            if (selfPlayerId && hotelPlayed.playerId === selfPlayerId) {
              setMovesLeft((current) => Math.max(0, current - 1));
            }

            setInitialGameState((current) => {
              if (!current || !hotelPlayed.propertySet) {
                return current;
              }

              const nextProperties = [...current.properties];
              const propertyIndex = nextProperties.findIndex((propertySet) => {
                return propertySet.propertySetId === hotelPlayed.propertySet?.propertySetId;
              });

              if (propertyIndex === -1) {
                nextProperties.push(hotelPlayed.propertySet);
              } else {
                nextProperties[propertyIndex] = hotelPlayed.propertySet;
              }

              const playedCardId = hotelPlayed.card?.cardId;
              const nextYourHand =
                selfPlayerId === hotelPlayed.playerId && playedCardId
                  ? {
                      ...current.yourHand,
                      cards:
                        current.yourHand?.cards.filter((card) => {
                          return card.cardId !== playedCardId;
                        }) ?? [],
                    }
                  : current.yourHand;

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== hotelPlayed.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                };
              });

              return {
                ...current,
                seqNum: hotelPlayed.seqNum,
                players: nextPlayers,
                properties: nextProperties,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== hotelPlayed.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                };
              });
            });
          }

          const playPassGoRes = message.monopolyDealMessage?.playPassGoRes;
          if (
            playPassGoRes &&
            selfPlayerId &&
            playPassGoRes.playerId === selfPlayerId
          ) {
            setMovesLeft((current) => Math.max(0, current - 1));

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const playedCardId = playPassGoRes.lastPlayedCard?.cardId;
              const nextYourHand = {
                ...current.yourHand,
                cards: [
                  ...(current.yourHand?.cards.filter((card) => {
                    return card.cardId !== playedCardId;
                  }) ?? []),
                  ...playPassGoRes.cards,
                ],
              };
              const handDelta = playPassGoRes.cards.length - 1;
              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== playPassGoRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards + handDelta),
                };
              });

              return {
                ...current,
                seqNum: playPassGoRes.seqNum,
                players: nextPlayers,
                yourHand: nextYourHand,
                lastAction: playPassGoRes.lastPlayedCard,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== playPassGoRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards + playPassGoRes.cards.length - 1),
                };
              });
            });
          }

          const playPassGoMaskedRes =
            message.monopolyDealMessage?.playPassGoMaskedRes;
          if (playPassGoMaskedRes) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const handDelta = playPassGoMaskedRes.numCards - 1;
              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== playPassGoMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards + handDelta),
                };
              });

              return {
                ...current,
                seqNum: playPassGoMaskedRes.seqNum,
                players: nextPlayers,
                lastAction: playPassGoMaskedRes.lastPlayedCard,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== playPassGoMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards + playPassGoMaskedRes.numCards - 1),
                };
              });
            });
          }

          const playActionRes = message.monopolyDealMessage?.playActionRes;
          if (playActionRes) {
            const isSelfPlay =
              !!selfPlayerId && playActionRes.playerId === selfPlayerId;
            if (isSelfPlay) {
              setMovesLeft((current) => Math.max(0, current - 1));
            }

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const playedCardId = playActionRes.lastPlayedCard?.cardId;
              const nextPlayers = current.players.map((player) => {
                if (!isSelfPlay || player.playerId !== playActionRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
                };
              });

              const nextYourHand =
                isSelfPlay && playedCardId
                  ? {
                      ...current.yourHand,
                      cards:
                        current.yourHand?.cards.filter((card) => {
                          return card.cardId !== playedCardId;
                        }) ?? [],
                    }
                  : current.yourHand;

              return {
                ...current,
                seqNum: playActionRes.seqNum,
                players: nextPlayers,
                yourHand: nextYourHand,
                lastAction: playActionRes.lastPlayedCard,
              };
            });

            if (isSelfPlay) {
              setPlayers((currentPlayers) => {
                return currentPlayers.map((player) => {
                  if (player.playerId !== playActionRes.playerId) {
                    return player;
                  }

                  return {
                    ...player,
                    handCards: Math.max(0, player.handCards - 1),
                  };
                });
              });
            }
          }

          const demandCreated = message.monopolyDealMessage?.demandCreated;
          if (demandCreated) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const incomingDemand = demandCreated.demand;
              if (!incomingDemand) {
                return {
                  ...current,
                  pendingRent: undefined,
                };
              }

              const nextDemands = current.demands.some((demand) => {
                return demand.id === incomingDemand.id;
              })
                ? current.demands.map((demand) => {
                    return demand.id === incomingDemand.id ? incomingDemand : demand;
                  })
                : [...current.demands, incomingDemand];

              return {
                ...current,
                demands: nextDemands,
                pendingRent: undefined,
              };
            });
          }

          const demandDenied = message.monopolyDealMessage?.demandDenied;
          if (demandDenied) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const hasMatchingDemand = current.demands.some((demand) => {
                return demand.id === demandDenied.demandId;
              });
              if (!hasMatchingDemand) {
                return current;
              }

              return {
                ...current,
                seqNum: demandDenied.seqNum,
                demands: current.demands.filter((demand) => {
                  return demand.id !== demandDenied.demandId;
                }),
              };
            });
          }

          const compliedDemand = message.monopolyDealMessage?.compliedDemand;
          if (compliedDemand) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              return {
                ...current,
                seqNum: compliedDemand.seqNum,
                demands: current.demands.filter((demand) => {
                  return demand.id !== compliedDemand.demandId;
                }),
              };
            });
          }

          const transferProperty = message.monopolyDealMessage?.transferProperty;
          if (transferProperty) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const sourceId = transferProperty.sourceId;
              const targetId = transferProperty.targetId;
              const nextSourceSets = transferProperty.sourcePropertySets.map((propertySet) => {
                return {
                  ...propertySet,
                  playerId: sourceId,
                };
              });
              const nextTargetSets = transferProperty.targetPropertySets.map((propertySet) => {
                return {
                  ...propertySet,
                  playerId: targetId,
                };
              });

              const sourceTransferredCardIds = new Set<string>();
              for (const propertySet of nextSourceSets) {
                for (const card of propertySet.cards) {
                  sourceTransferredCardIds.add(card.cardId);
                }
              }

              const targetTransferredCardIds = new Set<string>();
              for (const propertySet of nextTargetSets) {
                for (const card of propertySet.cards) {
                  targetTransferredCardIds.add(card.cardId);
                }
              }

              const prunedProperties = current.properties
                .map((propertySet) => {
                  if (propertySet.playerId === targetId && sourceTransferredCardIds.size > 0) {
                    return {
                      ...propertySet,
                      cards: propertySet.cards.filter(
                        (card) => !sourceTransferredCardIds.has(card.cardId),
                      ),
                    };
                  }

                  if (propertySet.playerId === sourceId && targetTransferredCardIds.size > 0) {
                    return {
                      ...propertySet,
                      cards: propertySet.cards.filter(
                        (card) => !targetTransferredCardIds.has(card.cardId),
                      ),
                    };
                  }

                  return propertySet;
                })
                .filter((propertySet) => propertySet.cards.length > 0);

              const upsertSetMap = new Map<string, (typeof prunedProperties)[number]>();
              for (const propertySet of prunedProperties) {
                upsertSetMap.set(propertySet.propertySetId, propertySet);
              }
              for (const propertySet of nextSourceSets) {
                upsertSetMap.set(propertySet.propertySetId, propertySet);
              }
              for (const propertySet of nextTargetSets) {
                upsertSetMap.set(propertySet.propertySetId, propertySet);
              }

              const nextProperties = Array.from(upsertSetMap.values());
              const nextPlayers = recalculatePlayerStatsFromBoard(
                current.players,
                current.money,
                nextProperties,
              );
              nextPlayersSnapshot = nextPlayers;

              return {
                ...current,
                seqNum: transferProperty.seqNum,
                properties: nextProperties,
                players: nextPlayers,
              };
            });

            if (nextPlayersSnapshot) {
              setPlayers(nextPlayersSnapshot);
            }
          }

          const transferPropertySet = message.monopolyDealMessage?.transferPropertySet;
          if (transferPropertySet) {
            let nextPlayersSnapshot: Player[] | null = null;
            setInitialGameState((current) => {
              if (!current || !transferPropertySet.propertySet) {
                return current;
              }

              const nextTransferredSet = {
                ...transferPropertySet.propertySet,
                playerId: transferPropertySet.sourceId,
              };

              const nextProperties = current.properties
                .filter((propertySet) => {
                  return propertySet.propertySetId !== transferPropertySet.propertySet?.propertySetId;
                })
                .filter((propertySet) => {
                  return propertySet.propertySetId !== nextTransferredSet.propertySetId;
                });

              nextProperties.push(nextTransferredSet);

              const nextPlayers = recalculatePlayerStatsFromBoard(
                current.players,
                current.money,
                nextProperties,
              );
              nextPlayersSnapshot = nextPlayers;

              return {
                ...current,
                seqNum: transferPropertySet.seqNum,
                properties: nextProperties,
                players: nextPlayers,
              };
            });

            if (nextPlayersSnapshot) {
              setPlayers(nextPlayersSnapshot);
            }
          }

          const pendingRentCreated =
            message.monopolyDealMessage?.pendingRentCreated;
          if (pendingRentCreated) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              return {
                ...current,
                pendingRent: pendingRentCreated.pendingRent,
                demands: [],
              };
            });
          }

          const pendingRentResolved =
            message.monopolyDealMessage?.pendingRentResolved;
          if (pendingRentResolved) {
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              return {
                ...current,
                seqNum: pendingRentResolved.seqNum,
                pendingRent: undefined,
              };
            });
          }

          const discardCardsRes = message.monopolyDealMessage?.discardCardsRes;
          if (discardCardsRes) {
            const discardedCardIds = new Set(
              discardCardsRes.cards.map((card) => card.cardId),
            );
            const discardedCount = discardCardsRes.cards.length;

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== discardCardsRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - discardedCount),
                };
              });

              const nextYourHand =
                discardCardsRes.playerId === selfPlayerId
                  ? {
                      ...current.yourHand,
                      cards:
                        current.yourHand?.cards.filter((card) => {
                          return !discardedCardIds.has(card.cardId);
                        }) ?? [],
                    }
                  : current.yourHand;

              return {
                ...current,
                seqNum: discardCardsRes.seqNum,
                players: nextPlayers,
                yourHand: nextYourHand,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== discardCardsRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - discardedCount),
                };
              });
            });

            if (discardCardsRes.playerId === selfPlayerId) {
              setSelectedDiscardCardIds(new Set());
            }
          }

          const discardCardsMaskedRes =
            message.monopolyDealMessage?.discardCardsMaskedRes;
          if (discardCardsMaskedRes) {
            const discardedCount = discardCardsMaskedRes.numCards;

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== discardCardsMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - discardedCount),
                };
              });

              return {
                ...current,
                seqNum: discardCardsMaskedRes.seqNum,
                players: nextPlayers,
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== discardCardsMaskedRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - discardedCount),
                };
              });
            });
          }

          const rearrangeCardRes = message.monopolyDealMessage?.rearrangeCardRes;
          if (rearrangeCardRes?.propertySet) {
            const movedCardId = rearrangeCardRes.card?.cardId;
            const targetPropertySet = rearrangeCardRes.propertySet;

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextProperties = current.properties
                .map((propertySet) => {
                  if (propertySet.playerId !== rearrangeCardRes.playerId) {
                    return propertySet;
                  }

                  return {
                    ...propertySet,
                    cards: movedCardId
                      ? propertySet.cards.filter((card) => card.cardId !== movedCardId)
                      : propertySet.cards,
                  };
                })
                .filter((propertySet) => {
                  return (
                    propertySet.playerId !== rearrangeCardRes.playerId ||
                    propertySet.cards.length > 0
                  );
                });

              const targetIndex = nextProperties.findIndex((propertySet) => {
                return propertySet.propertySetId === targetPropertySet.propertySetId;
              });

              if (targetIndex === -1) {
                nextProperties.push(targetPropertySet);
              } else {
                nextProperties[targetIndex] = targetPropertySet;
              }

              return {
                ...current,
                seqNum: rearrangeCardRes.seqNum,
                properties: nextProperties,
                players: current.players.map((player) => {
                  if (player.playerId !== rearrangeCardRes.playerId) {
                    return player;
                  }

                  return {
                    ...player,
                    completedSets: rearrangeCardRes.sets,
                  };
                }),
              };
            });

            setPlayers((currentPlayers) => {
              return currentPlayers.map((player) => {
                if (player.playerId !== rearrangeCardRes.playerId) {
                  return player;
                }

                return {
                  ...player,
                  completedSets: rearrangeCardRes.sets,
                };
              });
            });
          }

          const transferCards = message.monopolyDealMessage?.transferCards;
          if (transferCards) {
            const sourceId = transferCards.sourceId;
            const targetId = transferCards.targetId;
            const isSelfSource = !!selfPlayerId && selfPlayerId === sourceId;
            const {
              sourceCompletedSets,
              targetCompletedSets,
              sourceMoney,
              targetMoney,
            } = getTransferCardsPlayerStats(transferCards);
            const transferredMoneyCards = transferCards.cards;
            const transferredPropertySets = transferCards.propertySets;
            const moneyCardIds = toCardIdSet(transferredMoneyCards);
            const propertySetCardIds = toPropertySetCardIdSet(transferredPropertySets);
            const sourcePropertyCardIdsToRemove = new Set<string>([
              ...moneyCardIds,
              ...propertySetCardIds,
            ]);

            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const nextMoney = current.money.map((moneyPile) => {
                if (moneyPile.playerId === sourceId) {
                  return {
                    ...moneyPile,
                    cards: moneyPile.cards.filter(
                      (card) => !moneyCardIds.has(card.cardId),
                    ),
                  };
                }

                if (moneyPile.playerId === targetId) {
                  return {
                    ...moneyPile,
                    cards: [...moneyPile.cards, ...transferredMoneyCards],
                  };
                }

                return moneyPile;
              });

              const hasTargetMoneyPile = nextMoney.some(
                (moneyPile) => moneyPile.playerId === targetId,
              );
              if (!hasTargetMoneyPile && transferredMoneyCards.length > 0) {
                nextMoney.push({
                  playerId: targetId,
                  cards: transferredMoneyCards,
                });
              }

              const baseProperties = current.properties
                .map((propertySet) => {
                  if (propertySet.playerId !== sourceId) {
                    return propertySet;
                  }

                  return {
                    ...propertySet,
                    cards: propertySet.cards.filter(
                      (card) => !sourcePropertyCardIdsToRemove.has(card.cardId),
                    ),
                  };
                })
                .filter((propertySet) => propertySet.cards.length > 0);

              const targetPropertySets = transferredPropertySets.map((propertySet) => {
                return {
                  ...propertySet,
                  playerId: targetId,
                };
              });

              const nextProperties = [...baseProperties, ...targetPropertySets];
              const nextPlayers = applyTransferCardsPlayerStats(
                current.players,
                sourceId,
                targetId,
                sourceCompletedSets,
                targetCompletedSets,
                sourceMoney,
                targetMoney,
              );

              return {
                ...current,
                money: nextMoney,
                properties: nextProperties,
                players: nextPlayers,
                demands: isSelfSource
                  ? current.demands.filter((demand) => demand.sourceId !== sourceId)
                  : current.demands,
              };
            });

            setPlayers((currentPlayers) => {
              return applyTransferCardsPlayerStats(
                currentPlayers,
                sourceId,
                targetId,
                sourceCompletedSets,
                targetCompletedSets,
                sourceMoney,
                targetMoney,
              );
            });
          }

            console.log("[game-ws] message", toGameServerMessageJson(message));
          } catch (error) {
            console.error("[game-ws] failed to decode message", error);
            pushErrorNotice(toClientGameError(error, "WS_MESSAGE_DECODE_FAILED"));
          }
        })();
      };

      socket.onerror = (event) => {
        if (
          !isCurrentSession() ||
          socketRef.current !== socket ||
          globalGameSocket !== socket
        ) {
          return;
        }
        console.log("[game-ws] error", event);
      };

      socket.onclose = (event) => {
        console.log("[game-ws] close", {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        });

        if (!isCurrentSession()) {
          return;
        }

        if (socketRef.current === socket) {
          socketRef.current = null;
        }
        if (globalGameSocket === socket) {
          globalGameSocket = null;
        }

        const isDuplicateSocketClose =
          event.reason.toLowerCase().includes("duplicate socket created") ||
          event.reason.includes("API002");
        if (isDuplicateSocketClose) {
          console.log("[game-ws] duplicate socket close; skip reconnect");
          return;
        }

        scheduleReconnect();
      };
    };

    connectSocket();

    return () => {
      didUnmount = true;
      if (globalSocketSessionId === sessionId) {
        globalSocketSessionId = sessionId + 1;
      }
      clearReconnectTimer();
      if (socketRef.current === activeSocket) {
        socketRef.current = null;
      }
      if (globalGameSocket === activeSocket) {
        globalGameSocket = null;
      }
      activeSocket?.close();
    };
  }, [gameId, pushErrorNotice, selfPlayerId]);

  useEffect(() => {
    if (!isDiscardRequired) {
      setSelectedDiscardCardIds((current) => {
        if (current.size === 0) {
          return current;
        }

        return new Set();
      });
      return;
    }

    setSelectedDiscardCardIds((current) => {
      const validCardIds = new Set(handCards.map((card) => card.cardId));
      const next = new Set<string>();
      for (const cardId of current) {
        if (validCardIds.has(cardId)) {
          next.add(cardId);
        }
      }

       if (next.size === current.size) {
        let hasDifference = false;
        for (const cardId of next) {
          if (!current.has(cardId)) {
            hasDifference = true;
            break;
          }
        }

        if (!hasDifference) {
          return current;
        }
      }

      return next;
    });
  }, [handCardIdsKey, handCards, isDiscardRequired]);

  const handleSendChatMessage = useCallback((payload: string) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] chat send skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    sendGameChatMessage(socket, payload);
  }, [notifySocketUnavailable]);

  const canDragMoneyCard = useCallback(
    (card: Card) => {
      if (!selfPlayerId || !currentTurnPlayerId) {
        return false;
      }

      if (selfPlayerId !== currentTurnPlayerId) {
        return false;
      }

      if (movesLeft <= 0) {
        return false;
      }

      return (
        card.category === Category.CATEGORY_MONEY ||
        card.category === Category.CATEGORY_ACTION
      );
    },
    [currentTurnPlayerId, movesLeft, selfPlayerId],
  );

  const handlePlayMoneyCard = useCallback(
    (cardId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || !canDragMoneyCard(card)) {
        console.log("[game-ui] play money blocked by frontend checks", {
          cardId,
        });
        notifyFrontendRule(
          "You cannot play that card as money right now.",
          "PLAY_MONEY_FRONTEND_BLOCKED",
        );
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play money skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      sendGamePlayMoneyMessage(socket, cardId);
    },
    [canDragMoneyCard, initialGameState, notifyFrontendRule, notifySocketUnavailable],
  );

  const handlePlayPassGoCard = useCallback(
    (cardId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] action-pile play blocked by frontend checks", {
          cardId,
        });
        notifyFrontendRule(
          "You can only play action cards into the action pile.",
          "ACTION_PILE_INVALID_CARD",
        );
        notifyFrontendRule("Invalid action card.", "DEBT_COLLECTOR_INVALID_CARD");
        notifyFrontendRule("Invalid action card.", "SLY_DEAL_INVALID_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] action-pile play blocked; not your turn", {
          cardId,
        });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] action-pile play blocked; no moves left", {
          cardId,
        });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] action-pile play skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      if (card.assetKey === AssetKey.ASSET_KEY_PASS_GO) {
        console.log("[game-ui] play pass-go sent", { cardId });
        sendGamePlayPassGoMessage(socket, cardId);
        return;
      }

      if (card.assetKey === AssetKey.ASSET_KEY_ITS_MY_BIRTHDAY) {
        console.log("[game-ui] play-its-my-birthday sent", { cardId });
        sendGamePlayItsMyBirthdayMessage(socket, cardId);
        return;
      }

      if (isNormalRentAssetKey(card.assetKey)) {
        console.log("[game-ui] play-rent sent", { cardId });
        sendGamePlayRentMessage(socket, cardId);
        return;
      }

      console.log("[game-ui] action-pile play blocked; unsupported action", {
        cardId,
        assetKey: card.assetKey,
      });
      notifyFrontendRule("That action card cannot be played from this zone.", "UNSUPPORTED_ACTION_CARD");
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const canPlayPropertyCard = useCallback(
    (card: Card) => {
      if (!selfPlayerId || !currentTurnPlayerId) {
        return false;
      }

      if (selfPlayerId !== currentTurnPlayerId) {
        return false;
      }

      if (movesLeft <= 0) {
        return false;
      }

      const isPropertyCategory =
        card.category === Category.CATEGORY_PURE_PROPERTY ||
        card.category === Category.CATEGORY_WILD_PROPERTY;
      const isHouseOrHotelAction =
        card.category === Category.CATEGORY_ACTION &&
        (card.assetKey === AssetKey.ASSET_KEY_HOUSE ||
          card.assetKey === AssetKey.ASSET_KEY_HOTEL);

      return isPropertyCategory || isHouseOrHotelAction;
    },
    [currentTurnPlayerId, movesLeft, selfPlayerId],
  );

  const handlePlayPropertyCard = useCallback(
    (cardId: string, propertySetId?: string, activeColor?: Color) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || !canPlayPropertyCard(card)) {
        console.log("[game-ui] play property blocked by frontend checks", {
          cardId,
        });
        notifyFrontendRule(
          "You cannot play that property card right now.",
          "PLAY_PROPERTY_FRONTEND_BLOCKED",
        );
        return;
      }

      const isHouse = card.assetKey === AssetKey.ASSET_KEY_HOUSE;
      const isHotel = card.assetKey === AssetKey.ASSET_KEY_HOTEL;

      if (isHouse || isHotel) {
        if (!propertySetId) {
          pushErrorNotice(
            {
              message: "Choose one of your complete sets before playing this card.",
              code: "HOUSE_HOTEL_REQUIRES_SET",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        const targetSet = initialGameState?.properties.find((propertySet) => {
          return propertySet.propertySetId === propertySetId;
        });

        if (!targetSet || targetSet.playerId !== selfPlayerId) {
          pushErrorNotice(
            {
              message: "You can only play this card on one of your own complete sets.",
              code: "HOUSE_HOTEL_INVALID_TARGET_SET",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        if (
          targetSet.color === Color.COLOR_UTILITY ||
          targetSet.color === Color.COLOR_RAILROAD
        ) {
          pushErrorNotice(
            {
              message: "House and Hotel cards cannot be played on Utility or Railroad sets.",
              code: "HOUSE_HOTEL_INVALID_SET_COLOR",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        const isCompleteSet = isPropertySetCompleteForBuilding(targetSet);
        const hasHouse = targetSet.cards.some((setCard) => {
          return setCard.assetKey === AssetKey.ASSET_KEY_HOUSE;
        });
        const hasHotel = targetSet.cards.some((setCard) => {
          return setCard.assetKey === AssetKey.ASSET_KEY_HOTEL;
        });

        if (!isCompleteSet) {
          pushErrorNotice(
            {
              message: "House and Hotel cards can only be played on complete sets.",
              code: "HOUSE_HOTEL_REQUIRES_COMPLETE_SET",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        if (isHouse && hasHouse) {
          pushErrorNotice(
            {
              message: "This set already has a House.",
              code: "HOUSE_ALREADY_PRESENT",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        if (isHotel && !hasHouse) {
          pushErrorNotice(
            {
              message: "Hotel can only be played on a complete set that already has a House.",
              code: "HOTEL_REQUIRES_HOUSE",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }

        if (isHotel && hasHotel) {
          pushErrorNotice(
            {
              message: "This set already has a Hotel.",
              code: "HOTEL_ALREADY_PRESENT",
            },
            { title: "Cannot play card", eyebrow: "Frontend rule" },
          );
          return;
        }
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play property skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      if (isHouse && propertySetId) {
        sendGamePlayHouseMessage(socket, {
          cardId,
          propertySetId,
        });
        return;
      }

      if (isHotel && propertySetId) {
        sendGamePlayHotelMessage(socket, {
          cardId,
          propertySetId,
        });
        return;
      }

      sendGamePlayPropertyMessage(socket, {
        cardId,
        propertySetId,
        activeColor:
          activeColor ??
          (card.category === Category.CATEGORY_WILD_PROPERTY &&
          card.activeColor !== Color.COLOR_UNSPECIFIED
            ? card.activeColor
            : undefined),
      });
    },
    [
      canPlayPropertyCard,
      initialGameState,
      notifyFrontendRule,
      notifySocketUnavailable,
      pushErrorNotice,
      selfPlayerId,
    ],
  );

  const handleRearrangeCard = useCallback(
    (cardId: string, propertySetId?: string, color?: Color) => {
      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] rearrange blocked; not your turn", {
          cardId,
        });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      const ownPropertySets = initialGameState?.properties.filter((propertySet) => {
        return propertySet.playerId === selfPlayerId;
      });
      const card = ownPropertySets
        ?.flatMap((propertySet) => propertySet.cards)
        .find((candidate) => candidate.cardId === cardId);

      if (
        !card ||
        (card.category !== Category.CATEGORY_PURE_PROPERTY &&
          card.category !== Category.CATEGORY_WILD_PROPERTY)
      ) {
        console.log("[game-ui] rearrange blocked; invalid card", {
          cardId,
        });
        notifyFrontendRule("Only property cards can be rearranged.", "REARRANGE_INVALID_CARD");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] rearrange skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      sendGameRearrangeCardMessage(socket, {
        cardId,
        propertySetId,
        color,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePlayDebtCollectorCard = useCallback(
    (cardId: string, targetPlayerId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play debt collector blocked by frontend checks", {
          cardId,
          targetPlayerId,
        });
        return;
      }

      if (card.assetKey !== AssetKey.ASSET_KEY_DEBT_COLLECTOR) {
        console.log("[game-ui] play debt collector blocked; wrong card", {
          cardId,
          assetKey: card.assetKey,
        });
        notifyFrontendRule("That card is not Debt Collector.", "DEBT_COLLECTOR_WRONG_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] play debt collector blocked; not your turn", {
          cardId,
        });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] play debt collector blocked; no moves left", {
          cardId,
        });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      if (!targetPlayerId || targetPlayerId === selfPlayerId) {
        console.log("[game-ui] play debt collector blocked; invalid target", {
          cardId,
          targetPlayerId,
        });
        notifyFrontendRule("Choose a valid opponent target.", "INVALID_TARGET_PLAYER");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play debt collector skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      console.log("[game-ui] play debt collector sent", {
        cardId,
        targetPlayerId,
      });
      sendGamePlayDebtCollectorMessage(socket, {
        cardId,
        targetId: targetPlayerId,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePlayWildRentCard = useCallback(
    (cardId: string, targetPlayerId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play wild rent blocked by frontend checks", {
          cardId,
          targetPlayerId,
        });
        notifyFrontendRule("Invalid action card.", "WILD_RENT_INVALID_CARD");
        return;
      }

      if (card.assetKey !== AssetKey.ASSET_KEY_RENT_WILD) {
        console.log("[game-ui] play wild rent blocked; wrong card", {
          cardId,
          assetKey: card.assetKey,
        });
        notifyFrontendRule("That card is not Wild Rent.", "WILD_RENT_WRONG_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] play wild rent blocked; not your turn", {
          cardId,
        });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] play wild rent blocked; no moves left", {
          cardId,
        });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      if (!targetPlayerId || targetPlayerId === selfPlayerId) {
        console.log("[game-ui] play wild rent blocked; invalid target", {
          cardId,
          targetPlayerId,
        });
        notifyFrontendRule("Choose a valid opponent target.", "INVALID_TARGET_PLAYER");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play wild rent skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      console.log("[game-ui] play wild rent sent", {
        cardId,
        targetPlayerId,
      });
      sendGamePlayWildRentMessage(socket, {
        cardId,
        targetId: targetPlayerId,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePlaySlyDealCard = useCallback(
    (cardId: string, targetPlayerId: string, targetCardId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play sly deal blocked by frontend checks", {
          cardId,
          targetPlayerId,
          targetCardId,
        });
        return;
      }

      if (card.assetKey !== AssetKey.ASSET_KEY_SLY_DEAL) {
        console.log("[game-ui] play sly deal blocked; wrong card", {
          cardId,
          assetKey: card.assetKey,
        });
        notifyFrontendRule("That card is not Sly Deal.", "SLY_DEAL_WRONG_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] play sly deal blocked; not your turn", { cardId });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] play sly deal blocked; no moves left", { cardId });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      if (!targetPlayerId || targetPlayerId === selfPlayerId || !targetCardId) {
        console.log("[game-ui] play sly deal blocked; invalid target", {
          cardId,
          targetPlayerId,
          targetCardId,
        });
        notifyFrontendRule("Choose a valid target card from an opponent.", "SLY_DEAL_INVALID_TARGET");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play sly deal skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      console.log("[game-ui] play sly deal sent", {
        cardId,
        targetPlayerId,
        targetCardId,
      });
      sendGamePlaySlyDealMessage(socket, {
        cardId,
        targetId: targetPlayerId,
        targetCardId,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePlayForcedDealCard = useCallback(
    (
      cardId: string,
      targetPlayerId: string,
      sourceCardId: string,
      targetCardId: string,
    ) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play forced deal blocked by frontend checks", {
          cardId,
          targetPlayerId,
          sourceCardId,
          targetCardId,
        });
        notifyFrontendRule("Invalid action card.", "FORCED_DEAL_INVALID_CARD");
        return;
      }

      if (card.assetKey !== AssetKey.ASSET_KEY_FORCED_DEAL) {
        console.log("[game-ui] play forced deal blocked; wrong card", {
          cardId,
          assetKey: card.assetKey,
        });
        notifyFrontendRule("That card is not Forced Deal.", "FORCED_DEAL_WRONG_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] play forced deal blocked; not your turn", { cardId });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] play forced deal blocked; no moves left", { cardId });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      if (
        !targetPlayerId ||
        targetPlayerId === selfPlayerId ||
        !sourceCardId ||
        !targetCardId
      ) {
        console.log("[game-ui] play forced deal blocked; invalid selection", {
          cardId,
          targetPlayerId,
          sourceCardId,
          targetCardId,
        });
        notifyFrontendRule("Select one opponent card and one of your cards.", "FORCED_DEAL_INVALID_SELECTION");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play forced deal skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      console.log("[game-ui] play forced deal sent", {
        cardId,
        targetPlayerId,
        sourceCardId,
        targetCardId,
      });
      sendGamePlayForcedDealMessage(socket, {
        cardId,
        targetId: targetPlayerId,
        sourceCardId,
        targetCardId,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePlayDealBreakerCard = useCallback(
    (cardId: string, targetPlayerId: string, propertySetId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play deal breaker blocked by frontend checks", {
          cardId,
          targetPlayerId,
          propertySetId,
        });
        notifyFrontendRule("Invalid action card.", "DEAL_BREAKER_INVALID_CARD");
        return;
      }

      if (card.assetKey !== AssetKey.ASSET_KEY_DEAL_BREAKER) {
        console.log("[game-ui] play deal breaker blocked; wrong card", {
          cardId,
          assetKey: card.assetKey,
        });
        notifyFrontendRule("That card is not Deal Breaker.", "DEAL_BREAKER_WRONG_CARD");
        return;
      }

      if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
        console.log("[game-ui] play deal breaker blocked; not your turn", { cardId });
        notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
        return;
      }

      if (movesLeft <= 0) {
        console.log("[game-ui] play deal breaker blocked; no moves left", { cardId });
        notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
        return;
      }

      if (!targetPlayerId || targetPlayerId === selfPlayerId || !propertySetId) {
        console.log("[game-ui] play deal breaker blocked; invalid selection", {
          cardId,
          targetPlayerId,
          propertySetId,
        });
        notifyFrontendRule("Select a complete set from an opponent.", "DEAL_BREAKER_INVALID_SELECTION");
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play deal breaker skipped; socket not open");
        notifySocketUnavailable();
        return;
      }

      console.log("[game-ui] play deal breaker sent", {
        cardId,
        targetPlayerId,
        propertySetId,
      });
      sendGamePlayDealBreakerMessage(socket, {
        cardId,
        targetId: targetPlayerId,
        propertySetId,
      });
    },
    [
      currentTurnPlayerId,
      initialGameState,
      movesLeft,
      notifyFrontendRule,
      notifySocketUnavailable,
      selfPlayerId,
    ],
  );

  const handlePassTurn = useCallback(() => {
    if (selfPlayerId) {
      const selfPropertySets = (initialGameState?.properties ?? []).filter(
        (propertySet) => propertySet.playerId === selfPlayerId,
      );
      const incompleteSetCountByColor = new Map<Color, number>();

      for (const propertySet of selfPropertySets) {
        const requiredCount = minPropertyCountForCompleteSet(propertySet.color);
        if (!Number.isFinite(requiredCount)) {
          continue;
        }

        const propertyCardCount = propertySet.cards.filter((card) => {
          return isPropertyCardForSetCompletion(card);
        }).length;
        if (propertyCardCount >= requiredCount) {
          continue;
        }

        const nextCount =
          (incompleteSetCountByColor.get(propertySet.color) ?? 0) + 1;
        incompleteSetCountByColor.set(propertySet.color, nextCount);

        if (nextCount > 1) {
          notifyFrontendRule(
            "You cannot pass turn with multiple incomplete sets of the same color.",
            "INVALID_INCOMPLETE_PROPERTY_SETS",
          );
          return;
        }
      }
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] complete turn skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    sendGameCompleteTurnMessage(socket);
  }, [initialGameState?.properties, notifyFrontendRule, notifySocketUnavailable, selfPlayerId]);

  const handleComplyPaymentDemand = useCallback((demandId: string, cardIds: string[]) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] comply payment demand skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    console.log("[game-ui] comply payment demand sent", { demandId, cardIds });
    sendGameComplyPaymentDemandMessage(socket, {
      demandId,
      cardIds,
    });
  }, [notifySocketUnavailable]);

  const handleComplyPropertyDemand = useCallback((demandId: string) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] comply property demand skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    console.log("[game-ui] comply property demand sent", { demandId });
    sendGameComplyPropertyDemandMessage(socket, demandId);
  }, [notifySocketUnavailable]);

  const handleComplyPropertySetDemand = useCallback((demandId: string) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] comply property-set demand skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    console.log("[game-ui] comply property-set demand sent", { demandId });
    sendGameComplyPropertySetDemandMessage(socket, demandId);
  }, [notifySocketUnavailable]);

  const handleToggleDiscardCard = useCallback(
    (cardId: string) => {
      if (!isDiscardRequired) {
        return;
      }

      setSelectedDiscardCardIds((current) => {
        const next = new Set(current);
        if (next.has(cardId)) {
          next.delete(cardId);
        } else {
          next.add(cardId);
        }
        return next;
      });
    },
    [isDiscardRequired],
  );

  const handleSubmitDiscard = useCallback(() => {
    if (!isDiscardRequired) {
      return;
    }

    if (selectedDiscardCardIds.size !== discardRequiredCount) {
      console.log("[game-ui] discard blocked; invalid selected count", {
        selected: selectedDiscardCardIds.size,
        required: discardRequiredCount,
      });
      notifyFrontendRule(
        `Select exactly ${discardRequiredCount} card${discardRequiredCount === 1 ? "" : "s"} to discard.`,
        "DISCARD_INVALID_SELECTION_COUNT",
      );
      return;
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] discard skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    sendGameDiscardCardsMessage(socket, Array.from(selectedDiscardCardIds));
  }, [
    discardRequiredCount,
    isDiscardRequired,
    notifyFrontendRule,
    notifySocketUnavailable,
    selectedDiscardCardIds,
  ]);

  const handleDenyDemand = useCallback((demandId: string) => {
    const justSayNoCard = initialGameState?.yourHand?.cards.find((card) => {
      return card.assetKey === AssetKey.ASSET_KEY_JUST_SAY_NO;
    });
    if (!justSayNoCard) {
      console.log("[game-ui] deny demand blocked; no JUST_SAY_NO in hand");
      notifyFrontendRule("You need a Just Say No card to deny this demand.", "MISSING_JUST_SAY_NO");
      return;
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] deny demand skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    sendGameDenyDemandMessage(socket, {
      demandId,
      cardId: justSayNoCard.cardId,
    });
  }, [initialGameState, notifyFrontendRule, notifySocketUnavailable]);

  const handlePlayDoubleTheRentCard = useCallback(() => {
    const card = initialGameState?.yourHand?.cards.find((candidate) => {
      return candidate.assetKey === AssetKey.ASSET_KEY_DOUBLE_THE_RENT;
    });

    if (!card) {
      console.log("[game-ui] play double-the-rent blocked; card not in hand");
      notifyFrontendRule("Double The Rent is not in your hand.", "MISSING_DOUBLE_THE_RENT");
      return;
    }

    if (!selfPlayerId || !currentTurnPlayerId || selfPlayerId !== currentTurnPlayerId) {
      console.log("[game-ui] play double-the-rent blocked; not your turn", {
        cardId: card.cardId,
      });
      notifyFrontendRule("It is not your turn.", "NOT_YOUR_TURN");
      return;
    }

    if (movesLeft <= 0) {
      console.log("[game-ui] play double-the-rent blocked; no moves left", {
        cardId: card.cardId,
      });
      notifyFrontendRule("You have no moves left this turn.", "NO_MOVES_LEFT");
      return;
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] play double-the-rent skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    console.log("[game-ui] play double-the-rent sent", { cardId: card.cardId });
    sendGamePlayDoubleTheRentMessage(socket, card.cardId);
  }, [
    currentTurnPlayerId,
    initialGameState,
    movesLeft,
    notifyFrontendRule,
    notifySocketUnavailable,
    selfPlayerId,
  ]);

  const handleResolvePendingRent = useCallback(() => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] resolve pending rent skipped; socket not open");
      notifySocketUnavailable();
      return;
    }

    console.log("[game-ui] resolve pending rent sent");
    sendGameResolvePendingRentMessage(socket);
  }, [notifySocketUnavailable]);

  const winnerProfile = wonGame
    ? players.find((player) => player.playerId === wonGame.playerId)
    : null;
  const winnerName = wonGame
    ? wonGame.displayName ??
      winnerProfile?.displayName ??
      playerNameById[wonGame.playerId] ??
      wonGame.playerId
    : "";
  const winnerAvatarUrl = wonGame?.avatarUrl ?? winnerProfile?.avatarUrl ?? "";
  const winnerInitial = winnerName.trim().charAt(0).toUpperCase() || "W";

  return (
    <>
      <main
        className={wonGame ? "page game-page game-page--ended" : "page game-page"}
      >
        <section className="game-page__board">
          <MonopolyDealGameMount
            initialGameState={initialGameState}
            assetImageByKey={assetImageByKey}
            selfPlayerId={selfPlayerId ?? undefined}
            onPlayMoneyCard={handlePlayMoneyCard}
            onPlayPassGoCard={handlePlayPassGoCard}
            onPlayDebtCollectorCard={handlePlayDebtCollectorCard}
            onPlayWildRentCard={handlePlayWildRentCard}
            onPlaySlyDealCard={handlePlaySlyDealCard}
            onPlayDealBreakerCard={handlePlayDealBreakerCard}
            onPlayForcedDealCard={handlePlayForcedDealCard}
            onPlayDoubleTheRentCard={handlePlayDoubleTheRentCard}
            onResolvePendingRent={handleResolvePendingRent}
            onRearrangeCard={handleRearrangeCard}
            onDenyDemand={handleDenyDemand}
            isDiscardRequired={isDiscardRequired}
            requiredDiscardCount={discardRequiredCount}
            selectedDiscardCardIds={selectedDiscardCardIds}
            onToggleDiscardCard={handleToggleDiscardCard}
            onPlayPropertyCard={handlePlayPropertyCard}
            onComplyPaymentDemand={handleComplyPaymentDemand}
            onComplyPropertyDemand={handleComplyPropertyDemand}
            onComplyPropertySetDemand={handleComplyPropertySetDemand}
            onClientError={(error) => {
              pushErrorNotice(toClientGameError(error, "GAME_BOARD_RUNTIME"));
            }}
            onGameError={(error) => {
              pushErrorNotice(error, {
                title: "Cannot play card",
                eyebrow: "Frontend rule",
              });
            }}
          />
        </section>

        <aside className="game-sidebar" aria-label="Game sidebar">
          <ChatBox
            title="Game chat"
            messages={chatMessages}
            onSendMessage={handleSendChatMessage}
            getMessageKey={(message) => message.id}
            emptyMessage="No messages yet."
            renderMessage={(message) => {
              const author =
                playerNameById[message.playerId] ?? message.playerId;
              const authorAvatar =
                players.find((player) => player.playerId === message.playerId)?.avatarUrl ?? "";
              return (
                <article className="game-chat-line">
                  <img
                    className="game-chat-line__avatar"
                    src={authorAvatar}
                    alt={author}
                    loading="lazy"
                    referrerPolicy="no-referrer"
                  />
                  <p className="chat-message game-chat-line__message">
                    <span className="game-chat-line__author">{author}:</span> {message.text}
                  </p>
                </article>
              );
            }}
            className="game-chat-panel"
            messagesInnerClassName="game-chat-received-list"
          />

          <section className="game-sidebar-card game-players-card">
            <h2 className="game-sidebar-title">Players</h2>
            <div className="game-players-list">
              {players.length === 0 ? (
                <p className="game-sidebar-empty">Waiting for players</p>
              ) : (
                players.map((player) => (
                  <article
                    className="game-player-snippet"
                    key={player.playerId}
                  >
                    <img
                      className="game-player-avatar"
                      src={player.avatarUrl}
                      alt={player.displayName}
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                    <div className="game-player-meta">
                      <p className="game-player-name">
                        {player.displayName}
                        {player.playerId === currentTurnPlayerId ? (
                          <span className="game-player-pill">Current turn</span>
                        ) : null}
                      </p>
                      <p className="game-player-stats">
                        ${player.money} · {player.completedSets} sets · {player.handCards} cards
                      </p>
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>

          <TurnControlsCard
            onPassTurn={handlePassTurn}
            onSubmitDiscard={handleSubmitDiscard}
            movesLeft={movesLeft}
            showMovesLeft={selfPlayerId === currentTurnPlayerId}
            isDiscardRequired={isDiscardRequired}
            selectedDiscardCount={selectedDiscardCardIds.size}
            requiredDiscardCount={discardRequiredCount}
          />
        </aside>

        {wonGame ? (
          <section
            className="game-end-overlay"
            role="dialog"
            aria-modal="true"
            aria-labelledby="game-end-title"
          >
            <article className="game-end-modal">
              <p className="game-end-modal__eyebrow">Game over</p>
              <h2 id="game-end-title" className="game-end-modal__title">
                {winnerName} won the game
              </h2>
              <div className="game-end-modal__winner">
                {winnerAvatarUrl ? (
                  <img
                    className="game-end-modal__avatar"
                    src={winnerAvatarUrl}
                    alt={winnerName}
                    loading="lazy"
                    referrerPolicy="no-referrer"
                  />
                ) : (
                  <div className="game-end-modal__avatar game-end-modal__avatar--fallback" aria-hidden="true">
                    {winnerInitial}
                  </div>
                )}
                <p className="game-end-modal__winner-name">{winnerName}</p>
              </div>
              <dl className="game-end-modal__stats" aria-label="Winning stats">
                <div className="game-end-modal__stat">
                  <dt>Completed sets</dt>
                  <dd>{wonGame.numCompletedSets}</dd>
                </div>
                <div className="game-end-modal__stat">
                  <dt>Money</dt>
                  <dd>${wonGame.money}</dd>
                </div>
              </dl>
              <Button
                className="game-end-modal__button"
                onClick={() => {
                  navigate("/lobby");
                }}
              >
                Go Home
              </Button>
            </article>
          </section>
        ) : null}
      </main>

      <ErrorToastStack notices={errorNotices} onDismiss={dismissErrorNotice} />
    </>
  );
};

export default GamePage;
