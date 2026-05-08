<script lang="ts">
  // Small inline SVG icon set. Stroke icons in Heroicons / Lucide style.
  // Keeps bundle weight low and avoids a font/dep.
  export let name:
    | 'home'
    | 'home-fill'
    | 'dumbbell'
    | 'dumbbell-fill'
    | 'list'
    | 'list-fill'
    | 'clock'
    | 'clock-fill'
    | 'chart'
    | 'chart-fill'
    | 'play'
    | 'check'
    | 'plus'
    | 'chevron-right'
    | 'chevron-left'
    | 'sparkle'
    | 'cog'
    | 'cloud-off'
    | 'logout'
    | 'flame'
    | 'timer'
    | 'pencil'
    | 'trash';
  export let size: number = 24;
</script>

{#if name === 'home'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M3 11.5 12 4l9 7.5"/><path d="M5 10.5V20a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1v-9.5"/>
  </svg>
{:else if name === 'home-fill'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <path d="M11.36 3.42a1 1 0 0 1 1.28 0l8.36 6.96a1 1 0 0 1 .36.77V20a2 2 0 0 1-2 2h-3.5a.5.5 0 0 1-.5-.5V16a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v5.5a.5.5 0 0 1-.5.5H5a2 2 0 0 1-2-2v-8.85a1 1 0 0 1 .36-.77z"/>
  </svg>

{:else if name === 'dumbbell'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M6 8v8M3 10v4M9 6v12M15 6v12M18 8v8M21 10v4"/><path d="M9 12h6"/>
  </svg>
{:else if name === 'dumbbell-fill'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <rect x="2" y="9.5" width="2.5" height="5" rx="1"/>
    <rect x="5" y="7" width="3.5" height="10" rx="1"/>
    <rect x="9" y="11" width="6" height="2" rx="0.5"/>
    <rect x="15.5" y="7" width="3.5" height="10" rx="1"/>
    <rect x="19.5" y="9.5" width="2.5" height="5" rx="1"/>
  </svg>

{:else if name === 'list'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M9 6h12M9 12h12M9 18h12"/><circle cx="4" cy="6" r="1.5"/><circle cx="4" cy="12" r="1.5"/><circle cx="4" cy="18" r="1.5"/>
  </svg>
{:else if name === 'list-fill'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <circle cx="4" cy="6" r="2"/><circle cx="4" cy="12" r="2"/><circle cx="4" cy="18" r="2"/>
    <rect x="8.5" y="5" width="13" height="2" rx="1"/>
    <rect x="8.5" y="11" width="13" height="2" rx="1"/>
    <rect x="8.5" y="17" width="13" height="2" rx="1"/>
  </svg>

{:else if name === 'clock'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>
  </svg>
{:else if name === 'clock-fill'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <path d="M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18zm.6 4.4v4.16l2.7 1.8a.85.85 0 1 1-.94 1.41l-3.2-2.13a.85.85 0 0 1-.36-.7V7.4a.9.9 0 1 1 1.8 0z"/>
  </svg>

{:else if name === 'chart'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>
  </svg>
{:else if name === 'chart-fill'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <rect x="3" y="11" width="3.5" height="9" rx="1"/>
    <rect x="9" y="4" width="3.5" height="16" rx="1"/>
    <rect x="15" y="13" width="3.5" height="7" rx="1"/>
    <rect x="2" y="20" width="20" height="2" rx="1"/>
  </svg>

{:else if name === 'play'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <path d="M7 4.75a1 1 0 0 1 1.55-.83l11 7.25a1 1 0 0 1 0 1.66l-11 7.25A1 1 0 0 1 7 19.25z"/>
  </svg>

{:else if name === 'check'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M5 13l4 4L19 7"/>
  </svg>

{:else if name === 'plus'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true">
    <path d="M12 5v14M5 12h14"/>
  </svg>

{:else if name === 'chevron-right'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M9 6l6 6-6 6"/>
  </svg>

{:else if name === 'chevron-left'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M15 6l-6 6 6 6"/>
  </svg>

{:else if name === 'sparkle'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <path d="M12 2.5l1.95 4.55L18.5 9l-4.55 1.95L12 15.5l-1.95-4.55L5.5 9l4.55-1.95zM18.5 14.5l1.1 2.4 2.4 1.1-2.4 1.1-1.1 2.4-1.1-2.4-2.4-1.1 2.4-1.1z"/>
  </svg>

{:else if name === 'cog'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <circle cx="12" cy="12" r="3"/>
    <path d="M19.4 15a7.97 7.97 0 0 0 0-6l2-1.7-2-3.4-2.5 0.9a8 8 0 0 0-5.2-3l-0.5-2.6h-3.4l-0.5 2.6a8 8 0 0 0-5.2 3l-2.5-0.9-2 3.4 2 1.7a7.97 7.97 0 0 0 0 6l-2 1.7 2 3.4 2.5-0.9a8 8 0 0 0 5.2 3l0.5 2.6h3.4l0.5-2.6a8 8 0 0 0 5.2-3l2.5 0.9 2-3.4z" transform="scale(0.65) translate(6.5,6.5)"/>
  </svg>

{:else if name === 'cloud-off'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M3 3l18 18"/><path d="M19 17.6A4 4 0 0 0 18 10h-1.3a7 7 0 0 0-10.7-3.5"/><path d="M5.7 7.7A7 7 0 0 0 6 18h11"/>
  </svg>

{:else if name === 'logout'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
  </svg>

{:else if name === 'flame'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
    <path d="M12 2c1 4 5 5 5 10a5 5 0 0 1-10 0c0-2 1.5-3 1.5-5C8.5 5 10 3.5 12 2zm0 12a2 2 0 1 0 0 4 2 2 0 0 0 0-4z"/>
  </svg>

{:else if name === 'timer'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <circle cx="12" cy="13" r="8"/><path d="M12 9v4l2.5 2"/><path d="M9 2h6M12 5v3"/>
  </svg>

{:else if name === 'pencil'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M4 20h4l10-10-4-4L4 16v4z"/><path d="M14 6l4 4"/>
  </svg>

{:else if name === 'trash'}
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M4 7h16M9 7V4h6v3M6 7l1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13M10 11v7M14 11v7"/>
  </svg>
{/if}
