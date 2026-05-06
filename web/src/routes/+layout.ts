// Disable SSR for the entire app — we're a SPA served by the Go binary.
export const ssr = false;
export const prerender = false;
export const trailingSlash = 'never';
