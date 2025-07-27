export const getConfig = () => {
  if (window.appConfig?.VITE_API_BASE_URL) {
    return {
      API_BASE_URL: window.appConfig.VITE_API_BASE_URL,
    };
  }

  console.warn("Failed to fetch appConfig.js. Use default value.");
  return {
    API_BASE_URL: "http://localhost:8080",
  };
}; 