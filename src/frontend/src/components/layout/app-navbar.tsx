import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { logoutGoogleAuth } from "../../api/client";
import { getPlayer } from "../../api/player";
import {
  DEFAULT_THEME_ID,
  isThemeId,
  THEME_STORAGE_KEY,
  themeRegistry,
  type ThemeId,
} from "../../theme/registry";

const readThemeId = (): ThemeId => {
  if (typeof window === "undefined") {
    return DEFAULT_THEME_ID;
  }

  const savedTheme = window.localStorage.getItem(THEME_STORAGE_KEY);
  if (typeof savedTheme === "string" && isThemeId(savedTheme)) {
    return savedTheme;
  }

  return DEFAULT_THEME_ID;
};

const AppNavbar = () => {
  const navigate = useNavigate();
  const [themeId, setThemeId] = useState<ThemeId>(() => readThemeId());
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<{
    displayName: string;
    imageUrl: string;
  } | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    document.documentElement.dataset.theme = themeId;
    window.localStorage.setItem(THEME_STORAGE_KEY, themeId);
  }, [themeId]);

  useEffect(() => {
    let active = true;

    const fetchCurrentPlayer = async () => {
      const response = await getPlayer();
      if (!active || !response.ok) {
        return;
      }

      const player = (await response.json()) as {
        display_name?: string;
        image_url?: string;
      };

      if (!active || typeof player.image_url !== "string") {
        return;
      }

      setCurrentPlayer({
        displayName:
          typeof player.display_name === "string"
            ? player.display_name
            : "Player",
        imageUrl: player.image_url,
      });
    };

    void fetchCurrentPlayer();

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!isMenuOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setIsMenuOpen(false);
      }
    };

    window.addEventListener("mousedown", handlePointerDown);
    return () => {
      window.removeEventListener("mousedown", handlePointerDown);
    };
  }, [isMenuOpen]);

  const handleLogout = async () => {
    try {
      await logoutGoogleAuth();
    } finally {
      setCurrentPlayer(null);
      setIsMenuOpen(false);
      navigate("/login", { replace: true });
    }
  };

  return (
    <header className="app-nav">
      <div className="app-nav__inner">
        <div className="app-nav__left">
          <Link to="/lobby" className="app-nav__brand">
            Fun Kames!
          </Link>
        </div>

        <div className="app-nav__right">
          <select
            id="theme-select"
            className="field-input app-nav__theme-select"
            aria-label="Choose app theme"
            value={themeId}
            onChange={(event) => {
              const nextThemeId = event.target.value;
              if (isThemeId(nextThemeId)) {
                setThemeId(nextThemeId);
              }
            }}
          >
            {themeRegistry.map((theme) => (
              <option key={theme.id} value={theme.id}>
                {theme.label}
              </option>
            ))}
          </select>

          {currentPlayer ? (
            <div className="app-nav__profile" ref={menuRef}>
              <button
                type="button"
                className="app-nav__avatar-button"
                onClick={() => setIsMenuOpen((value) => !value)}
                aria-haspopup="menu"
                aria-expanded={isMenuOpen}
                aria-label="Open account menu"
              >
                <img
                  src={currentPlayer.imageUrl}
                  alt={currentPlayer.displayName}
                  className="app-nav__avatar"
                  referrerPolicy="no-referrer"
                />
              </button>

              {isMenuOpen ? (
                <div className="app-nav__menu" role="menu">
                  <button
                    type="button"
                    className="app-nav__menu-item"
                    onClick={() => {
                      setIsMenuOpen(false);
                      navigate("/profile");
                    }}
                    role="menuitem"
                  >
                    Profile
                  </button>
                  <button
                    type="button"
                    className="app-nav__menu-item"
                    onClick={() => {
                      setIsMenuOpen(false);
                      void handleLogout();
                    }}
                    role="menuitem"
                  >
                    Logout
                  </button>
                </div>
              ) : null}
            </div>
          ) : null}
        </div>
      </div>
    </header>
  );
};

export default AppNavbar;
