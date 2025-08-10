import { useRoutes } from "react-router-dom";
import routes from "./Routes";
import "./App.css";

function App() {
  const routing = useRoutes(routes);
  return (
    <TopMenu />
    {routing}
  );
}

export default App;
