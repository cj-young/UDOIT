import React, { createContext, useContext } from "react";

const ApiContext = createContext(null);

export default function ApiContextProvider({ children, api }) {
  return <ApiContext.Provider value={{ api }}>{children}</ApiContext.Provider>;
}

export function useApi() {
  const context = useContext(ApiContext);

  if (!context) {
    throw new Error("useApi must be used within ApiContextProvider");
  }

  return context.api;
}
