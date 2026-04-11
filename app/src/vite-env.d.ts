/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// Electron API exposed via preload script
interface Window {
  electronAPI?: {
    platform: string;
    isElectron: boolean;
    getApiUrl: () => string;
  };
}
