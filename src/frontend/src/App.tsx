import { Route, Routes } from "react-router-dom";
import AppNavbar from "./components/layout/app-navbar";
import LoginPage from "./pages/LoginPage";
import LobbyPage from "./pages/LobbyPage";
import ProfilePage from "./pages/ProfilePage";
import RoomPage from "./pages/RoomPage";
import ProtectedLobbyRoute from "./routes/ProtectedLobbyRoute";

const App = () => {
  return (
    <div className="app-shell">
      <AppNavbar />
      <div className="app-shell__content">
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedLobbyRoute />}>
            <Route path="/" element={<LobbyPage />} />
            <Route path="/lobby" element={<LobbyPage />} />
            <Route path="/profile" element={<ProfilePage />} />
            <Route path="/room/:room_id" element={<RoomPage />} />
          </Route>
        </Routes>
      </div>
    </div>
  );
};

export default App;
