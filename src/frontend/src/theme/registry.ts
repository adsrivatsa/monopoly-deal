export const THEME_STORAGE_KEY = "monopoly-deal-theme-id";

export const themeRegistry = [
  {
    id: "ocean-dark",
    label: "Ocean Dark",
  },
  {
    id: "lavender-light",
    label: "Lavender Light",
  },
] as const;

export type ThemeId = (typeof themeRegistry)[number]["id"];

export const DEFAULT_THEME_ID: ThemeId = "ocean-dark";

export const isThemeId = (value: string): value is ThemeId => {
  return themeRegistry.some((theme) => theme.id === value);
};
