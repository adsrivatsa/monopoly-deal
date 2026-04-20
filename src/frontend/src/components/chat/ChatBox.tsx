import { FormEvent, ReactNode, useEffect, useRef, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";

type ChatBoxProps<T> = {
  title: string;
  messages: T[];
  onSendMessage: (payload: string) => void;
  renderMessage: (message: T, index: number) => ReactNode;
  getMessageKey: (message: T, index: number) => string;
  emptyMessage?: string;
  placeholder?: string;
  maxLength?: number;
  className?: string;
  messagesInnerClassName?: string;
  stickToBottom?: boolean;
};

const ChatBox = <T,>({
  title,
  messages,
  onSendMessage,
  renderMessage,
  getMessageKey,
  emptyMessage = "No new events.",
  placeholder = "Type a message...",
  maxLength = 400,
  className,
  messagesInnerClassName,
  stickToBottom = true,
}: ChatBoxProps<T>) => {
  const [inputValue, setInputValue] = useState("");
  const messagesRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!stickToBottom || !messagesRef.current) {
      return;
    }

    messagesRef.current.scrollTop = messagesRef.current.scrollHeight;
  }, [messages, stickToBottom]);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const payload = inputValue.trim();
    if (!payload) {
      return;
    }

    onSendMessage(payload);
    setInputValue("");
  };

  return (
    <Card className={`room-chat-panel ${className ?? ""}`.trim()}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="room-chat-content">
        <div className="chat-log" role="log" aria-live="polite">
          <div
            ref={messagesRef}
            className={
              messagesInnerClassName
                ? `chat-log__messages ${messagesInnerClassName}`
                : "chat-log__messages"
            }
          >
            {messages.length === 0 ? (
              <p className="chat-message chat-message--empty">{emptyMessage}</p>
            ) : (
              messages.map((message, index) => {
                return (
                  <div className="chat-log__entry" key={getMessageKey(message, index)}>
                    {renderMessage(message, index)}
                  </div>
                );
              })
            )}
          </div>
        </div>

        <form className="chat-input-row" onSubmit={handleSubmit}>
          <input
            className="field-input"
            type="text"
            value={inputValue}
            onChange={(event) => {
              setInputValue(event.target.value);
            }}
            placeholder={placeholder}
            maxLength={maxLength}
          />
        </form>
      </CardContent>
    </Card>
  );
};

export default ChatBox;
