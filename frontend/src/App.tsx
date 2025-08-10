import { useState } from "react";
import { useRoutes } from "react-router-dom";
import "./App.css";
import TopMenu from "./components/TopMenu";
import routes from "./Routes";
import { searchContext } from "./utils/contexts";

function App() {
  const routing = useRoutes(routes);
  return (
    <TopMenu />
    {routing}
  );
}

export default App;
