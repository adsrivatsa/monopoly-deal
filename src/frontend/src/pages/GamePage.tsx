import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import {
  connectGameSocket,
  decodeGameServerMessage,
  sendGameChatMessage,
  sendGameCompleteTurnMessage,
  sendGamePlayMoneyMessage,
  sendGamePlayPassGoMessage,
  sendGamePlayPropertyMessage,
  toGameServerMessageJson,
} from "../api/gameSocket";
import { getPlayer } from "../api/player";
import ChatBox from "../components/chat/ChatBox";
import TurnControlsCard from "../components/game/TurnControlsCard";
import ErrorModal from "../components/ui/error-modal";
import MonopolyDealGameMount from "../game/monopoly_deal/MonopolyDealGameMount";
import {
  AssetKey,
  Category,
  Color,
  type Error as GameError,
  type AssetImage,
  type Card,
  type GameState,
  type Player,
} from "../generated/monopoly_deal";

type GameChatMessage = {
  id: string;
  playerId: string;
  text: string;
};

const toAssetImageMap = (assetImages: AssetImage[]): Record<number, string> => {
  return assetImages.reduce<Record<number, string>>((lookup, assetImage) => {
    if (assetImage.imageUrl) {
      lookup[assetImage.assetKey] = assetImage.imageUrl;
    }

    return lookup;
  }, {});
};

const GamePage = () => {
  const { game_id: gameId } = useParams();
  const socketRef = useRef<WebSocket | null>(null);
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
  const [modalError, setModalError] = useState<GameError | null>(null);

  useEffect(() => {
    void (async () => {
      const response = await getPlayer();
      if (!response.ok) {
        return;
      }

      const data = (await response.json()) as { player_id: string };
      setSelfPlayerId(data.player_id);
    })();
  }, []);

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
    const socket = connectGameSocket();
    socketRef.current = socket;

    console.log("[game-ws] connecting", socket.url);

    socket.onopen = () => {
      console.log("[game-ws] open", { gameId });
    };

    socket.onmessage = (event) => {
      void (async () => {
        try {
          const message = await decodeGameServerMessage(event.data);
          if (!message) {
            console.log("[game-ws] message (non-binary)", event.data);
            return;
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

          const gameError = message.monopolyDealMessage?.error;
          if (gameError) {
            setModalError(gameError);
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
                  cards: [playMoneyRes.card],
                });
              } else {
                const existingPile = nextMoney[moneyIndex];
                nextMoney[moneyIndex] = {
                  ...existingPile,
                  cards: [...existingPile.cards, playMoneyRes.card],
                };
              }

              const isSelfPlay = selfPlayerId === playMoneyRes.playerId;
              const nextYourHand = isSelfPlay
                ? {
                    ...current.yourHand,
                    cards:
                      current.yourHand?.cards.filter(
                        (card) => card.cardId !== playMoneyRes.card?.cardId,
                      ) ?? [],
                  }
                : current.yourHand;

              const nextPlayers = current.players.map((player) => {
                if (player.playerId !== playMoneyRes.playerId) {
                  return player;
                }

                return {
                  ...player,
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
            setInitialGameState((current) => {
              if (!current) {
                return current;
              }

              const playedCardId =
                playPropertyRes.propertySet.cards.at(-1)?.cardId;

              const nextProperties = [...current.properties];
              const propertyIndex = nextProperties.findIndex((propertySet) => {
                return (
                  propertySet.propertySetId ===
                  playPropertyRes.propertySet?.propertySetId
                );
              });

              if (propertyIndex === -1) {
                nextProperties.push(playPropertyRes.propertySet);
              } else {
                nextProperties[propertyIndex] = playPropertyRes.propertySet;
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

                return {
                  ...player,
                  handCards: Math.max(0, player.handCards - 1),
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

          console.log("[game-ws] message", toGameServerMessageJson(message));
        } catch (error) {
          console.error("[game-ws] failed to decode message", error);
        }
      })();
    };

    socket.onerror = (event) => {
      console.log("[game-ws] error", event);
    };

    socket.onclose = (event) => {
      console.log("[game-ws] close", {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
      });
    };

    return () => {
      socketRef.current = null;
      socket.close();
    };
  }, [gameId, selfPlayerId]);

  const handleSendChatMessage = useCallback((payload: string) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] chat send skipped; socket not open");
      return;
    }

    sendGameChatMessage(socket, payload);
  }, []);

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
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play money skipped; socket not open");
        return;
      }

      sendGamePlayMoneyMessage(socket, cardId);
    },
    [canDragMoneyCard, initialGameState],
  );

  const handlePlayPassGoCard = useCallback(
    (cardId: string) => {
      const card = initialGameState?.yourHand?.cards.find((candidate) => {
        return candidate.cardId === cardId;
      });

      if (!card || card.category !== Category.CATEGORY_ACTION) {
        console.log("[game-ui] play pass-go blocked by frontend checks", {
          cardId,
        });
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play pass-go skipped; socket not open");
        return;
      }

      console.log("[game-ui] play pass-go sent", { cardId });
      sendGamePlayPassGoMessage(socket, cardId);
    },
    [initialGameState],
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
        return;
      }

      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.log("[game-ws] play property skipped; socket not open");
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
    [canPlayPropertyCard, initialGameState],
  );

  const handlePassTurn = useCallback(() => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("[game-ws] complete turn skipped; socket not open");
      return;
    }

    sendGameCompleteTurnMessage(socket);
  }, []);

  return (
    <>
      <main className="page game-page">
        <section className="game-page__board">
          <MonopolyDealGameMount
            initialGameState={initialGameState}
            assetImageByKey={assetImageByKey}
            onPlayMoneyCard={handlePlayMoneyCard}
            onPlayPassGoCard={handlePlayPassGoCard}
            onPlayPropertyCard={handlePlayPropertyCard}
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
              return (
                <p className="chat-message">
                  <span className="chat-message__author">{author}:</span>{" "}
                  {message.text}
                </p>
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
                      <p className="game-player-name">{player.displayName}</p>
                      <p className="game-player-stats">
                        Money: {player.money} | Sets: {player.completedSets} |
                        Cards: {player.handCards}
                      </p>
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>

          <TurnControlsCard
            onPassTurn={handlePassTurn}
            movesLeft={movesLeft}
            showMovesLeft={selfPlayerId === currentTurnPlayerId}
          />
        </aside>
      </main>

      {modalError ? (
        <ErrorModal
          error={modalError}
          title="Game action failed"
          eyebrow="Game error"
          onClose={() => setModalError(null)}
        />
      ) : null}
    </>
  );
};

export default GamePage;
