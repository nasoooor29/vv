import { useNavigate } from "react-router";
import { useEffect } from "react";

function StorageIndex() {
  const navigate = useNavigate();

  useEffect(() => {
    navigate("./devices");
  }, [navigate]);

  return null;
}

export default StorageIndex;
