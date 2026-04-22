import { useNavigate } from "react-router-dom";
import Button from "../components/ui/button";

const HomePage = () => {
  const navigate = useNavigate();

  return (
    <main className="page">
      <h1>Home</h1>
      <Button
        onClick={() => {
          navigate("/lobby");
        }}
      >
        Start Playing!
      </Button>
    </main>
  );
};

export default HomePage;
