import { Link } from "react-router-dom";
import OAuthButtons from "../components/auth/OAuthButtons";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { appConfig } from "../config";

const LoginPage = () => {
  return (
    <main className="page">
      <Card className="auth-card">
        <CardHeader>
          <p className="eyebrow">Welcome back</p>
          <CardTitle>Sign in to Monopoly Deal</CardTitle>
          <CardDescription>
            Use Google OAuth to sign in and return to your lobby automatically.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <OAuthButtons
            providers={[
              {
                label: "Continue with Google",
                mark: "G",
                href: appConfig.auth.googleLoginUrl,
              },
            ]}
          />

          <div className="oauth-separator">
            <span>or</span>
          </div>

          <p className="auth-note">
            By continuing, you agree to the game rules and fair-play policy.
          </p>

          <Link to="/" className="text-link">
            Return to lobby preview
          </Link>
        </CardContent>
      </Card>
    </main>
  );
};

export default LoginPage;
