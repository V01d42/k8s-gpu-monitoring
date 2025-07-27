declare global {
  interface Window {
    appConfig?: {
      VITE_API_BASE_URL: string;
    };
  }
}

export {}; 