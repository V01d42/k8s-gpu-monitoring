import { useState } from "react";
import { useRoutes } from "react-router-dom";
import "./App.css";
import TopMenu from "./components/TopMenu";
import routes from "./Routes";
import { searchContext } from "./utils/contexts";

function App() {
  const routing = useRoutes(routes);
  const [searchText, setSearchText] = useState<string>('');

  return (
  <searchContext.Provider value={{searchText, setSearchText}}>
    <TopMenu />
    {routing}
  </searchContext.Provider>
  );
}

export default App;
