import type { RefObject } from "react";
import ChatBox from "../../../components/chat/ChatBox";
import type { Player } from "../../../generated/deal_no_mercy";
import type { DealNoMercyBoardViewMode } from "../DealNoMercyGameMount";
import type { ActionHistoryEntry, GameChatMessage } from "./dealNoMercyPageTypes";

type DealNoMercyGameSidebarProps = {
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
  boardViewMode: DealNoMercyBoardViewMode;
  onBoardViewModeChange: (viewMode: DealNoMercyBoardViewMode) => void;
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

// Kinds whose payload is a list of card asset keys rendered as a CardRow.
const CARD_ROW_KINDS: ReadonlySet<ActionHistoryEntry["kind"]> = new Set([
  "playAction",
  "discardCards",
  "paymentComplied",
  "distribution",
  "pickpocket",
  "propertySteal",
  "propertySetSteal",
  "bankSwap",
]);

// Kinds whose payload is a single card asset key rendered as a CardImage.
const CARD_IMAGE_KINDS: ReadonlySet<ActionHistoryEntry["kind"]> = new Set([
  "playMoney",
  "playProperty",
  "playShack",
  "bigPayday",
  "demandsCreated",
  "debtSettled",
]);

const ActionHistoryEntryView = ({
  entry,
  assetImageByKey,
}: ActionHistoryEntryViewProps) => {
  if (entry.kind === "turnDivider") {
    return (
      <div className="game-action-history-divider" aria-hidden="true">
        {entry.text}
      </div>
    );
  }

  if (entry.kind === "startTurn" || entry.kind === "maskedStartTurn") {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          {entry.text}
          {typeof entry.drawCount === "number" ? ` (${entry.drawCount})` : ""}
        </p>
      </article>
    );
  }

  if (CARD_ROW_KINDS.has(entry.kind) && (entry.cardAssetKeys?.length ?? 0) > 0) {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          {entry.text}
        </p>
        <CardRow
          assetKeys={entry.cardAssetKeys ?? []}
          assetImageByKey={assetImageByKey}
          alt="Card"
          entryId={entry.id}
        />
      </article>
    );
  }

  if (CARD_IMAGE_KINDS.has(entry.kind) && typeof entry.cardAssetKey === "number") {
    return (
      <article className="game-action-history-line game-action-history-line--event">
        <p className="chat-message game-chat-line__message game-action-history-text">
          {entry.text}
        </p>
        <div className="game-action-history-cards-row">
          <CardImage
            assetKey={entry.cardAssetKey}
            assetImageByKey={assetImageByKey}
            alt="Card"
          />
        </div>
      </article>
    );
  }

  // Fallback for "generic", "denied", "goAgain", "debtTrap", and any entry that
  // lacks the card payload its kind would otherwise carry. Render the label only.
  return (
    <p className="chat-message game-chat-line__message game-action-history-line">
      {entry.text}
    </p>
  );
};

const DealNoMercyGameSidebar = ({
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
}: DealNoMercyGameSidebarProps) => {
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

export default DealNoMercyGameSidebar;
