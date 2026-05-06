<script lang="ts">
  import { page } from '$app/stores';
  import Icon from './Icon.svelte';

  type Tab = {
    href: string;
    label: string;
    icon: 'home' | 'dumbbell' | 'list' | 'clock' | 'chart';
    iconFill: 'home-fill' | 'dumbbell-fill' | 'list-fill' | 'clock-fill' | 'chart-fill';
    match: (path: string) => boolean;
  };

  const tabs: Tab[] = [
    { href: '/', label: 'Today', icon: 'home', iconFill: 'home-fill', match: (p) => p === '/' },
    { href: '/exercises', label: 'Exercises', icon: 'dumbbell', iconFill: 'dumbbell-fill', match: (p) => p.startsWith('/exercises') || p.startsWith('/machines') },
    { href: '/routines', label: 'Routines', icon: 'list', iconFill: 'list-fill', match: (p) => p.startsWith('/routines') },
    { href: '/history', label: 'History', icon: 'clock', iconFill: 'clock-fill', match: (p) => p === '/history' },
    { href: '/reports', label: 'Reports', icon: 'chart', iconFill: 'chart-fill', match: (p) => p === '/reports' }
  ];

  $: path = $page.url.pathname;
</script>

<nav class="bottom-nav">
  {#each tabs as tab}
    {@const active = tab.match(path)}
    <a href={tab.href} class:active>
      <Icon name={active ? tab.iconFill : tab.icon} size={24} />
      <span>{tab.label}</span>
    </a>
  {/each}
</nav>
