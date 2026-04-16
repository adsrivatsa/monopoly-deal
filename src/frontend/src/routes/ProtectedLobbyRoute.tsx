import { useEffect, useState } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { getPlayer } from "../api/player";

const ProtectedLobbyRoute = () => {
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    let active = true;

    const checkProfile = async () => {
      try {
        const response = await getPlayer();
        if (active && response.status === 200) {
          setIsAuthorized(true);
        }
      } finally {
        if (active) {
          setIsChecking(false);
        }
      }
    };

    void checkProfile();

    return () => {
      active = false;
    };
  }, []);

  if (isChecking) {
    return (
      <main className="page">
        <p className="lede">Checking your session...</p>
      </main>
    );
  }

  if (!isAuthorized) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
};

export default ProtectedLobbyRoute;
