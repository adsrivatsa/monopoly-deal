import { MouseEvent } from "react";
import Button from "./button";

type ErrorModalProps = {
  error: {
    message: string;
    code: string;
  };
  title?: string;
  eyebrow?: string;
  onClose: () => void;
};

const ErrorModal = ({
  error,
  title = "Something went wrong",
  eyebrow = "Request failed",
  onClose,
}: ErrorModalProps) => {
  const onCardClick = (event: MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();
  };

  return (
    <div className="error-modal-backdrop" onClick={onClose} role="presentation">
      <div
        className="error-modal-card"
        onClick={onCardClick}
        role="dialog"
        aria-modal="true"
        aria-labelledby="error-modal-title"
      >
        <p className="eyebrow">{eyebrow}</p>
        <h2 id="error-modal-title" className="error-modal-title">
          {title}
        </h2>
        <p className="error-modal-message">{error.message}</p>
        <p className="error-modal-code">Code: {error.code}</p>
        <Button onClick={onClose}>Close</Button>
      </div>
    </div>
  );
};

export default ErrorModal;
