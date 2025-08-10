import { createContext } from "react";

export const searchContext = createContext({
  searchText: "",
  setSearchText: (() => {}) as (v: string) => void,
});
