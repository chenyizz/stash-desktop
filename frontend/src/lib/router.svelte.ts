// Minimal hash router. See docs/ARCHITECTURE.md ADR-006.
// Routes: "#/" (list) and "#/scenes/<id>" (detail).

export type RouteName = "list" | "detail";

export type Route = {
  name: RouteName;
  id: number;
};

function parseHash(hash: string): Route {
  const path = hash.startsWith("#") ? hash.slice(1) : hash;
  const match = /^\/scenes\/(\d+)$/.exec(path);
  if (match) {
    return { name: "detail", id: Number(match[1]) };
  }
  return { name: "list", id: 0 };
}

export const route = $state<Route>({ name: "list", id: 0 });

function sync() {
  const next = parseHash(window.location.hash);
  route.name = next.name;
  route.id = next.id;
}

export function navigate(path: string): void {
  const target = path.startsWith("#") ? path : "#" + path;
  if (window.location.hash === target) {
    sync();
    return;
  }
  window.location.hash = target;
}

export function initRouter(): () => void {
  sync();
  window.addEventListener("hashchange", sync);
  return () => window.removeEventListener("hashchange", sync);
}
