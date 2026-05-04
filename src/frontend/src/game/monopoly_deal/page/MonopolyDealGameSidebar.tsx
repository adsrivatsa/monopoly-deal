import type { RefObject } from "react";
import ChatBox from "../../../components/chat/ChatBox";
import type { Player } from "../../../generated/monopoly_deal";
import type { MonopolyDealBoardViewMode } from "../MonopolyDealGameMount";
import type { ActionHistoryEntry, GameChatMessage } from "./monopolyDealPageTypes";

type MonopolyDealGameSidebarProps = {
  isSidebarCollapsed: boolean;
  onToggleSidebar: () => void;
  actionHistoryListRef: RefObject<HTMLDivElement | null>;
  actionHistoryEntries: ActionHistoryEntry[];
  assetImageByKey: Record<number, string>;
  players: Player[];
  playerNameById: Record<string, string>;
  selfPlayerId: string | null;
  chatMessages: GameChatMessage[];
  onSendChatMessage: (payload: string) => void;
  boardViewMode: MonopolyDealBoardViewMode;
  onBoardViewModeChange: (viewMode: MonopolyDealBoardViewMode) => void;
};

type ActionHistoryEntryViewProps = {
  entry: ActionHistoryEntry;
  assetImageByKey: Record<number, string>;
  players: Player[];
  playerNameById: Record<string, string>;
  selfPlayerId: string | null;
};

const getPlayerAvatarUrl = (players: Player[], playerId: string): string => {
  return players.find((player) => player.playerId === playerId)?.avatarUrl ?? "";
};

const getPlayerName = (
  playerNameById: Record<string, string>,
  playerId: string,
  selfPlayerId: string | null,
  selfLabel: "You" | "you" | null,
): string => {
  if (selfLabel && selfPlayerId === playerId) {
    return selfLabel;
  }

  return playerNameById[playerId] ?? playerId;
};

const PlayerAvatar = ({
  playerId,
  players,
  playerNameById,
}: {
  playerId: string;
  players: Player[];
  playerNameById: Record<string, string>;
}) => (
  <img
    className="game-action-history-inline-avatar"
    src={getPlayerAvatarUrl(players, playerId)}
    alt={playerNameById[playerId] ?? playerId}
    loading="lazy"
    referrerPolicy="no-referrer"
  />
);

const CardImage = ({
  assetKey,
  assetImageByKey,
  alt,
}: {
  assetKey?: number;
  assetImageByKey: Record<number, string>;
  alt: string;
}) => (
  <img
    className="game-action-history-inline-card"
    src={typeof assetKey === "number" ? (assetImageByKey[assetKey] ?? "") : ""}
    alt={alt}
    loading="lazy"
    referrerPolicy="no-referrer"
  />
);

const CardRow = ({
  assetKeys,
  assetImageByKey,
  alt,
  entryId,
  ariaLabel,
}: {
  assetKeys: number[];
  assetImageByKey: Record<number, string>;
  alt: string;
  entryId: string;
  ariaLabel?: string;
}) => (
  <div className="game-action-history-cards-row" aria-label={ariaLabel}>
    {assetKeys.map((assetKey, index) => (
      <CardImage
        assetKey={assetKey}
        assetImageByKey={assetImageByKey}
        alt={alt}
        key={`${entryId}-${assetKey}-${index}`}
      />
    ))}
  </div>
);

const ActionActor = ({
  playerId,
  players,
  playerNameById,
  selfPlayerId,
  selfLabel = "You",
}: {
  playerId: string;
  players: Player[];
  playerNameById: Record<string, string>;
  selfPlayerId: string | null;
  selfLabel?: "You" | "you" | null;
}) => (
  <>
    <PlayerAvatar playerId={playerId} players={players} playerNameById={playerNameById} />
    <span className="game-chat-line__author">
      {getPlayerName(playerNameById, playerId, selfPlayerId, selfLabel)}
    </span>
  </>
);

const ActionHistoryEntryView = ({
  entry,
  assetImageByKey,
  players,
  playerNameById,
  selfPlayerId,
}: ActionHistoryEntryViewProps) => {
  if (entry.kind === "turnDivider") {
    return <hr className="game-action-history-divider" aria-hidden="true" />;
  }

  if (entry.kind === "paymentComplied" && entry.playerId) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          paid{" "}
          {entry.targetPlayerId ? (
            <ActionActor
              playerId={entry.targetPlayerId}
              players={players}
              playerNameById={playerNameById}
              selfPlayerId={selfPlayerId}
              selfLabel="you"
            />
          ) : (
            <span className="game-chat-line__author">them</span>
          )}
          {(entry.cardAssetKeys?.length ?? 0) > 0 ? ":" : " nothing."}
        </p>
        {(entry.cardAssetKeys?.length ?? 0) > 0 ? (
          <CardRow
            assetKeys={entry.cardAssetKeys ?? []}
            assetImageByKey={assetImageByKey}
            alt="Paid card"
            entryId={entry.id}
          />
        ) : null}
      </article>
    );
  }

  if (entry.kind === "discardCards" && entry.playerId && (entry.cardAssetKeys?.length ?? 0) > 0) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          discarded {entry.cardAssetKeys?.length ?? 0}{" "}
          {(entry.cardAssetKeys?.length ?? 0) === 1 ? "card:" : "cards:"}
        </p>
        <CardRow
          assetKeys={entry.cardAssetKeys ?? []}
          assetImageByKey={assetImageByKey}
          alt="Discarded card"
          entryId={entry.id}
        />
      </article>
    );
  }

  if (entry.kind === "maskedDiscardCards" && entry.playerId && (entry.drawCount ?? 0) > 0) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          discarded {entry.drawCount ?? 0} {(entry.drawCount ?? 0) === 1 ? "card." : "cards."}
        </p>
      </article>
    );
  }

  if (entry.kind === "propertySwap" && entry.sourcePlayerId && entry.targetPlayerId) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.sourcePlayerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          swapped cards with{" "}
          <ActionActor
            playerId={entry.targetPlayerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
            selfLabel="you"
          />
        </p>
        <div className="game-action-history-cards-row">
          <CardImage
            assetKey={entry.sourceCardAssetKey}
            assetImageByKey={assetImageByKey}
            alt="Swapped source card"
          />
          <span className="game-action-history-swap-arrows" aria-hidden="true">→ ←</span>
          <CardImage
            assetKey={entry.targetCardAssetKey}
            assetImageByKey={assetImageByKey}
            alt="Swapped target card"
          />
        </div>
      </article>
    );
  }

  if (
    (entry.kind === "propertySetSteal" || entry.kind === "propertySteal") &&
    entry.sourcePlayerId &&
    entry.targetPlayerId
  ) {
    const isSetSteal = entry.kind === "propertySetSteal";
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.sourcePlayerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          stole {isSetSteal ? "a set" : "a card"} from{" "}
          <ActionActor
            playerId={entry.targetPlayerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
            selfLabel="you"
          />
          .
        </p>
        <div className="game-action-history-cards-row">
          {isSetSteal ? (
            (entry.cardAssetKeys ?? []).map((assetKey, index) => (
              <CardImage
                assetKey={assetKey}
                assetImageByKey={assetImageByKey}
                alt="Stolen set card"
                key={`${entry.id}-${assetKey}-${index}`}
              />
            ))
          ) : (
            <CardImage
              assetKey={entry.cardAssetKey}
              assetImageByKey={assetImageByKey}
              alt="Stolen property card"
            />
          )}
        </div>
      </article>
    );
  }

  if (entry.kind === "playedCards" && entry.playerId) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          played:
        </p>
        <CardRow
          assetKeys={entry.cardAssetKeys ?? []}
          assetImageByKey={assetImageByKey}
          alt="Played card"
          entryId={entry.id}
        />
      </article>
    );
  }

  if (entry.kind === "startTurn" && entry.playerId) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
          />{" "}
          drew
        </p>
        <CardRow
          assetKeys={entry.cardAssetKeys ?? []}
          assetImageByKey={assetImageByKey}
          alt="Drawn card"
          entryId={entry.id}
          ariaLabel="Drawn cards"
        />
      </article>
    );
  }

  if (entry.kind === "maskedStartTurn" && entry.playerId) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
            selfLabel={null}
          />{" "}
          drew {entry.drawCount ?? 0} cards.
        </p>
      </article>
    );
  }

  if (
    (entry.kind === "playMoney" ||
      entry.kind === "playProperty" ||
      entry.kind === "demandsCreated") &&
    entry.playerId
  ) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          <ActionActor
            playerId={entry.playerId}
            players={players}
            playerNameById={playerNameById}
            selfPlayerId={selfPlayerId}
            selfLabel={null}
          />{" "}
          {entry.kind === "playMoney"
            ? "played money:"
            : entry.kind === "playProperty"
              ? "played property:"
              : "played action:"}
        </p>
        <div className="game-action-history-cards-row">
          <CardImage
            assetKey={entry.cardAssetKey}
            assetImageByKey={assetImageByKey}
            alt={
              entry.kind === "playMoney"
                ? "Money card"
                : entry.kind === "playProperty"
                  ? "Property card"
                  : "Action card"
            }
          />
        </div>
      </article>
    );
  }

  return (
    <p className="chat-message game-chat-line__message game-action-history-line">
      {entry.text}
    </p>
  );
};

const MonopolyDealGameSidebar = ({
  isSidebarCollapsed,
  onToggleSidebar,
  actionHistoryListRef,
  actionHistoryEntries,
  assetImageByKey,
  players,
  playerNameById,
  selfPlayerId,
  chatMessages,
  onSendChatMessage,
  boardViewMode,
  onBoardViewModeChange,
}: MonopolyDealGameSidebarProps) => {
  return (
    <aside className="game-sidebar" aria-label="Game sidebar">
      <section
        className={[
          "game-sidebar-card",
          "game-sidebar-panels",
          "game-sidebar-collapsible",
          isSidebarCollapsed ? "is-collapsed" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        onClick={(event) => {
          const target = event.target as HTMLElement | null;
          if (target?.closest(".game-sidebar-panels__content")) {
            return;
          }

          onToggleSidebar();
        }}
      >
        <div className="game-sidebar-card__header">
          <h2 className="game-sidebar-title">Game Panels</h2>
          <button
            type="button"
            className="game-collapse-button"
            onClick={(event) => {
              event.stopPropagation();
              onToggleSidebar();
            }}
            aria-expanded={!isSidebarCollapsed}
            aria-label={isSidebarCollapsed ? "Expand sidebar panels" : "Collapse sidebar panels"}
            title={isSidebarCollapsed ? "Expand" : "Collapse"}
          >
            <span className="game-collapse-icon game-collapse-icon--horizontal" aria-hidden="true">
              {isSidebarCollapsed ? "<" : ">"}
            </span>
            <span className="game-collapse-icon game-collapse-icon--vertical" aria-hidden="true">
              {isSidebarCollapsed ? "▾" : "▴"}
            </span>
          </button>
        </div>
        <div
          className={[
            "game-sidebar-panels__content",
            isSidebarCollapsed ? "is-collapsed" : "",
          ]
            .filter(Boolean)
            .join(" ")}
          aria-hidden={isSidebarCollapsed}
        >
          <section className="game-sidebar-card game-action-history-panel">
            <h2 className="game-sidebar-title">Action history</h2>
            <div
              className="game-action-history-list"
              role="log"
              aria-live="polite"
              ref={actionHistoryListRef}
            >
              {actionHistoryEntries.length === 0 ? (
                <p className="game-sidebar-empty">No actions yet.</p>
              ) : (
                actionHistoryEntries.map((entry) => (
                  <ActionHistoryEntryView
                    entry={entry}
                    assetImageByKey={assetImageByKey}
                    players={players}
                    playerNameById={playerNameById}
                    selfPlayerId={selfPlayerId}
                    key={entry.id}
                  />
                ))
              )}
            </div>
          </section>

          <ChatBox
            title="Game chat"
            messages={chatMessages}
            onSendMessage={onSendChatMessage}
            getMessageKey={(message) => message.id}
            emptyMessage="No messages yet."
            renderMessage={(message) => {
              const author = playerNameById[message.playerId] ?? message.playerId;
              const authorAvatar = getPlayerAvatarUrl(players, message.playerId);
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

          <section className="game-sidebar-card game-board-view-card">
            <h2 className="game-sidebar-title">Board view</h2>
            <button
              type="button"
              className="game-board-view-switch"
              onClick={() => {
                onBoardViewModeChange(
                  boardViewMode === "expanded" ? "compact" : "expanded",
                );
              }}
            >
              {boardViewMode === "expanded"
                ? "Switch to Compact view"
                : "Switch to Expanded view"}
            </button>
          </section>
        </div>
      </section>
    </aside>
  );
};

export default MonopolyDealGameSidebar;
