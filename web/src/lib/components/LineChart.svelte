<script lang="ts">
  // Multi-series SVG line chart. Each series gets its own color; lines are
  // drawn between sequential points within a series, and single-point series
  // render as a lone dot. No deps.

  import { onMount, onDestroy } from 'svelte';
  import type { Series } from './LineChart';

  export let series: Series[] = [];
  export let yLabel: string = '';
  export let height: number = 240;
  export let yFormat: (v: number) => string = (v) => v.toFixed(0);
  // Optional explicit x-axis window — useful when the user picks a fixed
  // range (3M / 6M / 12M) so the axis stays stable as data fills in.
  export let xRange: [number, number] | null = null;

  const PAD_L = 44;
  const PAD_R = 12;
  const PAD_T = 10;
  const PAD_B = 28;

  let width = 600;
  let svgEl: SVGSVGElement;
  let ro: ResizeObserver | null = null;

  function onResize() {
    if (svgEl?.parentElement) {
      width = Math.max(280, svgEl.parentElement.clientWidth);
    }
  }
  onMount(() => {
    onResize();
    if (typeof ResizeObserver !== 'undefined' && svgEl?.parentElement) {
      ro = new ResizeObserver(onResize);
      ro.observe(svgEl.parentElement);
    }
  });
  onDestroy(() => ro?.disconnect());

  $: allPoints = series.flatMap((s) => s.points);
  $: xs = allPoints.map((p) => p.x);
  $: ys = allPoints.map((p) => p.y);

  $: xMin = xRange ? xRange[0] : xs.length ? Math.min(...xs) : 0;
  $: xMax = xRange ? xRange[1] : xs.length ? Math.max(...xs) : 1;
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

  function pathFor(s: Series): string {
    const sorted = [...s.points].sort((a, b) => a.x - b.x);
    return sorted
      .map((p, i) => `${i === 0 ? 'M' : 'L'}${sx(p.x).toFixed(1)},${sy(p.y).toFixed(1)}`)
      .join(' ');
  }

  // Y-axis ticks (4 evenly-spaced, "nice" rounded numbers).
  $: yTicks = (() => {
    const span = yMax - yMin;
    if (span <= 0) return [yMin];
    const step = niceStep(span / 3);
    const start = Math.ceil(yMin / step) * step;
    const ticks: number[] = [];
    for (let v = start; v <= yMax + 1e-9; v += step) ticks.push(v);
    return ticks;
  })();

  function niceStep(raw: number): number {
    if (raw <= 0) return 1;
    const exp = Math.pow(10, Math.floor(Math.log10(raw)));
    const f = raw / exp;
    let nice: number;
    if (f < 1.5) nice = 1;
    else if (f < 3) nice = 2;
    else if (f < 7) nice = 5;
    else nice = 10;
    return nice * exp;
  }

  // X-axis ticks: roughly 4 across the visible range.
  $: xTicks = (() => {
    if (xMax === xMin) return [xMin];
    const n = 4;
    const step = (xMax - xMin) / (n - 1);
    return Array.from({ length: n }, (_, i) => xMin + step * i);
  })();

  function fmtDate(ms: number): string {
    return new Date(ms).toLocaleDateString([], { month: 'short', day: 'numeric' });
  }

  $: hasData = allPoints.length > 0;
</script>

{#if series.length > 0}
  <div class="legend">
    {#each series as s}
      <span class="legend-item">
        <span class="dot" style="background: {s.color};"></span>
        {s.name}
      </span>
    {/each}
  </div>
{/if}

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

  {#if yLabel}
    <text x={4} y={PAD_T + 6} font-size="10" fill="var(--muted)">{yLabel}</text>
  {/if}

  {#each xTicks as t}
    <text x={sx(t)} y={height - 8} text-anchor="middle" font-size="10" fill="var(--muted)">
      {fmtDate(t)}
    </text>
  {/each}

  <!-- one path per series -->
  {#each series as s}
    {#if s.points.length > 1}
      <path
        d={pathFor(s)}
        fill="none"
        stroke={s.color}
        stroke-width="2"
        stroke-linejoin="round"
        stroke-linecap="round"
      />
    {/if}
    {#each s.points as p}
      <circle cx={sx(p.x)} cy={sy(p.y)} r="3.5" fill={s.color}>
        {#if p.label}<title>{s.name}: {p.label}</title>{/if}
      </circle>
    {/each}
  {/each}

  {#if !hasData}
    <text x={width / 2} y={height / 2} text-anchor="middle" font-size="12" fill="var(--muted)">
      No data in this range
    </text>
  {/if}
</svg>

<style>
  svg { display: block; max-width: 100%; }
  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem 0.85rem;
    margin: 0 0 0.5rem;
    font-size: 0.85rem;
    color: var(--text-2);
  }
  .legend-item { display: inline-flex; align-items: center; gap: 0.35rem; }
  .dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }
</style>
