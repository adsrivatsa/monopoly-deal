import { useEffect } from "react";

type ErrorToastNotice = {
  id: string;
  message: string;
  code?: string;
  title?: string;
  eyebrow?: string;
  durationMs?: number;
};

type ErrorToastStackProps = {
  notices: ErrorToastNotice[];
  onDismiss: (id: string) => void;
};

const ErrorToastItem = ({
  notice,
  onDismiss,
}: {
  notice: ErrorToastNotice;
  onDismiss: (id: string) => void;
}) => {
  useEffect(() => {
    const timeoutMs = notice.durationMs ?? 5500;
    const timeoutId = window.setTimeout(() => {
      onDismiss(notice.id);
    }, timeoutMs);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [notice.durationMs, notice.id, onDismiss]);

  return (
    <article className="error-toast" role="status">
      <p className="error-toast__eyebrow">{notice.eyebrow ?? "Game error"}</p>
      <p className="error-toast__message">{notice.message}</p>
      {notice.code ? <p className="error-toast__code">Code: {notice.code}</p> : null}
      <button
        type="button"
        className="error-toast__close"
        onClick={() => onDismiss(notice.id)}
      >
        Dismiss
      </button>
    </article>
  );
};

const ErrorToastStack = ({ notices, onDismiss }: ErrorToastStackProps) => {
  if (notices.length === 0) {
    return null;
  }

  return (
    <section className="error-toast-stack" aria-live="polite" aria-label="Errors">
      {notices.map((notice) => (
        <ErrorToastItem key={notice.id} notice={notice} onDismiss={onDismiss} />
      ))}
    </section>
  );
};

export type { ErrorToastNotice };
export default ErrorToastStack;
