export interface Route {
  name: string;
  path: string;
}

export const ROUTES = {
  HOME: {
    name: "home",
    path: "/",
  },
  TALK_TO_DATA: {
    name: "talk-to-data",
    path: "/talk-to-data",
  },
} as const satisfies Record<string, Route>;
