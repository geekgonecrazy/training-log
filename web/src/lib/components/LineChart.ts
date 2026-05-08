// Type companion for LineChart.svelte. Svelte 4 won't let us re-export TS
// types from a .svelte module, so the shape lives here.

export interface Series {
  name: string;
  color: string;
  points: { x: number; y: number; label?: string }[];
}
