import Button from "../ui/button";

type OAuthProvider = {
  label: string;
  mark: string;
  href: string;
};

type OAuthButtonsProps = { providers: OAuthProvider[] };

const OAuthButtons = ({ providers }: OAuthButtonsProps) => {
  return (
    <div className="oauth-stack">
      {providers.map((provider) => (
        <a key={provider.href} href={provider.href}>
          <Button variant="outline" className="oauth-button">
            <span className="oauth-button__mark" aria-hidden="true">
              {provider.mark}
            </span>
            {provider.label}
          </Button>
        </a>
      ))}
    </div>
  );
};

export default OAuthButtons;
