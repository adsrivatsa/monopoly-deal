type DiscardPromptOverlayProps = {
  requiredDiscardCount: number;
  selectedDiscardCount: number;
};

const DiscardPromptOverlay = ({
  requiredDiscardCount,
  selectedDiscardCount,
}: DiscardPromptOverlayProps) => {
  return (
    <aside
      className="md-demand md-demand--discard"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">Discard required</p>
      <h3 className="md-demand__title">Choose cards to discard</h3>
      <p className="md-demand__line">
        Selected: {selectedDiscardCount}/{requiredDiscardCount}
      </p>
      <p className="md-demand__line">Select cards from your hand, then submit.</p>
    </aside>
  );
};

export default DiscardPromptOverlay;
