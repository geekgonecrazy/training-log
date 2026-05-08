<script lang="ts">
  // Tiny SVG line chart. Inputs are an array of {x: number (timestamp ms), y: number}.
  // No deps — handles axis range, padding, gridlines, point markers, missing data.

  export let points: { x: number; y: number; label?: string }[] = [];
  export let yLabel: string = '';
  export let height: number = 220;
  export let yFormat: (v: number) => string = (v) => v.toFixed(0);

  const PAD_L = 44;
  const PAD_R = 12;
  const PAD_T = 10;
  const PAD_B = 28;

  let width = 600;
  let svgEl: SVGSVGElement;

  // Track container width for responsive sizing.
  function onResize() {
    if (svgEl?.parentElement) {
      width = Math.max(280, svgEl.parentElement.clientWidth);
    }
  }

  import { onMount, onDestroy } from 'svelte';
  let ro: ResizeObserver | null = null;
  onMount(() => {
    onResize();
    if (typeof ResizeObserver !== 'undefined' && svgEl?.parentElement) {
      ro = new ResizeObserver(onResize);
      ro.observe(svgEl.parentElement);
    }
  });
  onDestroy(() => ro?.disconnect());

  $: sorted = [...points].sort((a, b) => a.x - b.x);
  $: xs = sorted.map((p) => p.x);
  $: ys = sorted.map((p) => p.y);
  $: xMin = xs.length ? Math.min(...xs) : 0;
  $: xMax = xs.length ? Math.max(...xs) : 1;
  $: yMin = ys.length ? Math.min(...ys, 0) : 0;
  $: yMaxRaw = ys.length ? Math.max(...ys) : 1;
  $: yMax = yMaxRaw === yMin ? yMin + 1 : yMaxRaw;

  $: plotW = width - PAD_L - PAD_R;
  $: plotH = height - PAD_T - PAD_B;

  function sx(x: number): number {
    if (xMax === xMin) return PAD_L + plotW / 2;
    return PAD_L + ((x - xMin) / (xMax - xMin)) * plotW;
  }
  function sy(y: number): number {
    if (yMax === yMin) return PAD_T + plotH / 2;
    return PAD_T + plotH - ((y - yMin) / (yMax - yMin)) * plotH;
  }

  $: pathD = sorted
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${sx(p.x).toFixed(1)},${sy(p.y).toFixed(1)}`)
    .join(' ');

  // Y-axis ticks (4 evenly-spaced).
  $: yTicks = (() => {
    const ticks: number[] = [];
    for (let i = 0; i <= 3; i++) {
      ticks.push(yMin + ((yMax - yMin) * i) / 3);
    }
    return ticks;
  })();

  // X-axis ticks: roughly 4 across the range.
  $: xTicks = (() => {
    if (sorted.length <= 1) return sorted.map((p) => p.x);
    const n = Math.min(4, sorted.length);
    const step = (xMax - xMin) / (n - 1);
    return Array.from({ length: n }, (_, i) => xMin + step * i);
  })();

  function fmtDate(ms: number): string {
    const d = new Date(ms);
    return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
  }
</script>

<svg
  bind:this={svgEl}
  viewBox="0 0 {width} {height}"
  width={width}
  {height}
  preserveAspectRatio="none"
  role="img"
  aria-label="line chart"
>
  <!-- gridlines + y labels -->
  {#each yTicks as t}
    <line
      x1={PAD_L}
      x2={width - PAD_R}
      y1={sy(t)}
      y2={sy(t)}
      stroke="rgba(255,255,255,0.06)"
      stroke-dasharray="2 4"
    />
    <text x={PAD_L - 6} y={sy(t) + 3} text-anchor="end" font-size="10" fill="var(--muted)">
      {yFormat(t)}
    </text>
  {/each}

  <!-- y axis label -->
  {#if yLabel}
    <text x={4} y={PAD_T + 6} font-size="10" fill="var(--muted)">{yLabel}</text>
  {/if}

  <!-- x labels -->
  {#each xTicks as t}
    <text x={sx(t)} y={height - 8} text-anchor="middle" font-size="10" fill="var(--muted)">
      {fmtDate(t)}
    </text>
  {/each}

  <!-- the line -->
  {#if sorted.length > 1}
    <path d={pathD} fill="none" stroke="var(--accent)" stroke-width="2" stroke-linejoin="round" stroke-linecap="round" />
  {/if}

  <!-- points -->
  {#each sorted as p}
    <circle cx={sx(p.x)} cy={sy(p.y)} r="3.5" fill="var(--accent)">
      {#if p.label}<title>{p.label}</title>{/if}
    </circle>
  {/each}

  {#if sorted.length === 0}
    <text x={width / 2} y={height / 2} text-anchor="middle" font-size="12" fill="var(--muted)">
      No data yet
    </text>
  {/if}
</svg>

<style>
  svg { display: block; max-width: 100%; }
</style>
