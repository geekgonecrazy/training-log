<script lang="ts">
  // Multi-series SVG line chart. Each series gets its own color; lines are
  // drawn between sequential points within a series, and single-point series
  // render as a lone dot. No deps.
  //
  // Click a legend entry to isolate that series — only it will render and
  // the y-axis rescales to its values. Click again (or click another) to
  // toggle. Parents bind `isolatedName` to control this externally.

  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import type { Series } from './LineChart';

  export let series: Series[] = [];
  export let yLabel: string = '';
  export let height: number = 240;
  export let yFormat: (v: number) => string = (v) => v.toFixed(0);
  // Optional explicit x-axis window — useful when the user picks a fixed
  // range (3M / 6M / 12M) so the axis stays stable as data fills in.
  export let xRange: [number, number] | null = null;
  // When set, only the named series is plotted; all other legend items are
  // dimmed. null = show everything.
  export let isolatedName: string | null = null;

  const dispatch = createEventDispatcher<{ legendClick: { name: string } }>();

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

  // Visible series: when isolated, only the named one is drawn. Y-axis is
  // computed from visible points so isolation actually zooms in.
  $: visibleSeries = isolatedName
    ? series.filter((s) => s.name === isolatedName)
    : series;
  $: visiblePoints = visibleSeries.flatMap((s) => s.points);
  $: ys = visiblePoints.map((p) => p.y);
  $: allXs = series.flatMap((s) => s.points).map((p) => p.x);

  $: xMin = xRange ? xRange[0] : allXs.length ? Math.min(...allXs) : 0;
  $: xMax = xRange ? xRange[1] : allXs.length ? Math.max(...allXs) : 1;
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

  $: hasData = visiblePoints.length > 0;

  function isDimmed(name: string): boolean {
    return isolatedName !== null && isolatedName !== name;
  }
</script>

{#if series.length > 0}
  <div class="legend">
    {#each series as s}
      <button
        type="button"
        class="legend-item"
        class:dimmed={isDimmed(s.name)}
        on:click={() => dispatch('legendClick', { name: s.name })}
        aria-pressed={isolatedName === s.name}
      >
        <span class="dot" style="background: {s.color};"></span>
        {s.name}
      </button>
    {/each}
    {#if isolatedName}
      <button
        type="button"
        class="legend-item show-all"
        on:click={() => dispatch('legendClick', { name: isolatedName ?? '' })}
      >Show all</button>
    {/if}
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

  <!-- one path per visible series -->
  {#each visibleSeries as s}
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
    gap: 0.4rem 0.6rem;
    margin: 0 0 0.5rem;
    font-size: 0.85rem;
    color: var(--text-2);
  }
  .legend-item {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 999px;
    padding: 0.2rem 0.55rem 0.2rem 0.4rem;
    color: inherit;
    font-size: inherit;
    font-weight: 500;
    min-height: 0;
    cursor: pointer;
    transition: background 0.12s ease, border-color 0.12s ease, opacity 0.12s ease;
  }
  .legend-item:hover { background: var(--surface-2); }
  .legend-item[aria-pressed='true'] {
    background: var(--surface-2);
    border-color: var(--hairline-strong);
  }
  .legend-item.dimmed { opacity: 0.35; }
  .legend-item.dimmed:hover { opacity: 0.7; }

  .legend-item.show-all {
    color: var(--accent);
    border-color: rgba(0, 210, 168, 0.3);
    padding: 0.2rem 0.7rem;
    font-weight: 600;
  }
  .legend-item.show-all:hover { background: rgba(0, 210, 168, 0.08); }

  .dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }
</style>
