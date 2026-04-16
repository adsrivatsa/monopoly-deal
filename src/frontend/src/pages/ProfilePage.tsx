import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getPlayer, updatePlayer } from "../api/player";

type Profile = {
  displayName: string;
  email: string;
  imageUrl: string;
};

const ProfilePage = () => {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [isEditingName, setIsEditingName] = useState(false);
  const [nameDraft, setNameDraft] = useState("");

  const submitUpdatedName = async (nextName: string): Promise<boolean> => {
    const response = await updatePlayer({
      display_name: nextName,
    });

    if (!response.ok) {
      if (response.status === 401) {
        navigate("/login", { replace: true });
      }
      return false;
    }

    return true;
  };

  useEffect(() => {
    let active = true;

    const fetchProfile = async () => {
      const response = await getPlayer();
      if (!active) {
        return;
      }

      if (!response.ok) {
        navigate("/login", { replace: true });
        return;
      }

      const payload = (await response.json()) as {
        display_name?: string;
        email?: string;
        image_url?: string;
      };

      setProfile({
        displayName:
          typeof payload.display_name === "string"
            ? payload.display_name
            : "Unknown Player",
        email: typeof payload.email === "string" ? payload.email : "-",
        imageUrl: typeof payload.image_url === "string" ? payload.image_url : "",
      });

      setNameDraft(
        typeof payload.display_name === "string"
          ? payload.display_name
          : "Unknown Player",
      );
    };

    void fetchProfile();

    return () => {
      active = false;
    };
  }, [navigate]);

  return (
    <main className="page profile-page">
      <section className="profile-card">
        <div className="profile-card__avatar-wrap">
          {profile?.imageUrl ? (
            <img
              src={profile.imageUrl}
              alt={profile.displayName}
              className="profile-card__avatar"
              referrerPolicy="no-referrer"
            />
          ) : (
            <div className="profile-card__avatar profile-card__avatar--placeholder" />
          )}
        </div>

        <div className="profile-card__details">
          {isEditingName ? (
            <form
              className="profile-name-form"
              onSubmit={(event) => {
                event.preventDefault();
                const nextName = nameDraft.trim();
                if (!nextName) {
                  return;
                }

                void (async () => {
                  const didUpdate = await submitUpdatedName(nextName);
                  if (!didUpdate) {
                    return;
                  }

                  setProfile((currentProfile) => {
                    if (!currentProfile) {
                      return currentProfile;
                    }

                    return {
                      ...currentProfile,
                      displayName: nextName,
                    };
                  });
                  setIsEditingName(false);
                })();
              }}
            >
              <input
                className="field-input profile-name-input"
                value={nameDraft}
                onChange={(event) => setNameDraft(event.target.value)}
                aria-label="Edit display name"
                autoFocus
              />
            </form>
          ) : (
            <div className="profile-name-row">
              <h1 className="profile-card__name">{profile?.displayName ?? "Loading..."}</h1>
              <button
                type="button"
                className="profile-name-edit"
                onClick={() => setIsEditingName(true)}
                aria-label="Edit display name"
              >
                ✎
              </button>
            </div>
          )}
          <p className="profile-card__email">{profile?.email ?? ""}</p>
        </div>
      </section>
    </main>
  );
};

export default ProfilePage;
