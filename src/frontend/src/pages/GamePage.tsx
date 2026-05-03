import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { getGame } from "../api/room";
import { Game } from "../api/models";

const GamePage = () => {
  const { game_id: gameId } = useParams();
  const navigate = useNavigate();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    const redirectToGameRoute = async () => {
      setErrorMessage(null);
      const result = await getGame();

      if (!active) {
        return;
      }

      if (!result.ok) {
        if (result.isTokenError) {
          navigate("/login", { replace: true });
          return;
        }

        if (result.isNotFound) {
          navigate("/lobby", { replace: true });
          return;
        }

        setErrorMessage(result.error?.message ?? "Unable to load game.");
        return;
      }

      const targetGameId = result.data.game_id || gameId;
      if (!targetGameId) {
        setErrorMessage("Unable to load game.");
        return;
      }

      switch (result.data.game) {
        case Game.MonopolyDeal:
          navigate(`/deal/${targetGameId}`, { replace: true });
          return;
        default:
          setErrorMessage(`Unsupported game: ${result.data.game}`);
      }
    };

    void redirectToGameRoute();

    return () => {
      active = false;
    };
  }, [gameId, navigate]);

  return (
    <main className="page">
      <h1>Game</h1>
      <p>{errorMessage ?? "Loading game..."}</p>
    </main>
  );
};

export default GamePage;
