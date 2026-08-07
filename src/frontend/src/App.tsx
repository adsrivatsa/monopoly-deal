import { Suspense, lazy } from "react";
import { Route, Routes } from "react-router-dom";
import AppNavbar from "./components/layout/app-navbar";
import ProtectedLobbyRoute from "./routes/ProtectedLobbyRoute";

const HomePage = lazy(() => import("./pages/HomePage"));
const LoginPage = lazy(() => import("./pages/LoginPage"));
const LobbyPage = lazy(() => import("./pages/LobbyPage"));
const ProfilePage = lazy(() => import("./pages/ProfilePage"));
const RoomPage = lazy(() => import("./pages/RoomPage"));
const GamePage = lazy(() => import("./pages/GamePage"));
const MonopolyDealGamePage = lazy(() => import("./game/monopoly_deal/page/MonopolyDealGamePage"));
const DealNoMercyGamePage = lazy(() => import("./game/deal_no_mercy/page/DealNoMercyGamePage"));

const App = () => {
  return (
    <div className="app-shell">
      <AppNavbar />
      <div className="app-shell__content">
        <Suspense fallback={<div className="page">Loading...</div>}>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route element={<ProtectedLobbyRoute />}>
              <Route path="/lobby" element={<LobbyPage />} />
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/room/:room_id" element={<RoomPage />} />
              <Route path="/game/:game_id" element={<GamePage />} />
              <Route path="/deal/:game_id" element={<MonopolyDealGamePage />} />
              <Route path="/no-mercy/:game_id" element={<DealNoMercyGamePage />} />
            </Route>
          </Routes>
        </Suspense>
      </div>
    </div>
  );
};

export default App;
