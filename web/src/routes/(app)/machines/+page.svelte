<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import { machines as api } from '$lib/api/endpoints';
  import type { Machine } from '$lib/api/types';

  let items: Machine[] = [];
  let loading = true;
  let editing: Machine | null = null;
  let creating = false;

  let formName = '';
  let formLocation = '';
  let formNotes = '';

  onMount(async () => {
    await refresh();
    loading = false;
  });

  async function refresh() {
    const r = await api.list();
    items = r.machines ?? [];
  }

  function startCreate() {
    creating = true;
    editing = null;
    formName = '';
    formLocation = '';
    formNotes = '';
  }

  function startEdit(m: Machine) {
    editing = m;
    creating = false;
    formName = m.name;
    formLocation = m.location;
    formNotes = m.notes;
  }

  async function save() {
    if (creating) {
      await api.create({ name: formName, location: formLocation, notes: formNotes });
    } else if (editing) {
      await api.update(editing.id, { name: formName, location: formLocation, notes: formNotes });
    }
    creating = false;
    editing = null;
    await refresh();
  }

  async function remove(m: Machine) {
    if (!confirm(`Delete "${m.name}"?`)) return;
    await api.delete(m.id);
    await refresh();
  }
</script>

<div class="app-shell">
  <AppHeader title="Machines" back="/exercises" />

  {#if !creating && !editing}
    <button class="primary" on:click={startCreate}>+ New machine</button>
  {:else}
    <div class="card stack">
      <div class="field">
        <label>Name</label>
        <input bind:value={formName} required />
      </div>
      <div class="field">
        <label>Location</label>
        <input bind:value={formLocation} />
      </div>
      <div class="field">
        <label>Notes</label>
        <textarea bind:value={formNotes} rows="2"></textarea>
      </div>
      <div class="row">
        <button class="primary" on:click={save}>Save</button>
        <button
          class="ghost"
          on:click={() => {
            creating = false;
            editing = null;
          }}
        >
          Cancel
        </button>
      </div>
    </div>
  {/if}

  <div style="margin-top: 1rem;">
    {#if loading}
      <p class="muted">Loading…</p>
    {:else if items.length === 0 && !creating}
      <p class="muted">No machines yet.</p>
    {:else}
      <div class="stack">
        {#each items as m}
          <div class="card row">
            <div class="grow">
              <strong>{m.name}</strong>
              {#if m.location}
                <div class="muted" style="font-size: 0.85rem;">{m.location}</div>
              {/if}
            </div>
            <button class="ghost" on:click={() => startEdit(m)}>Edit</button>
            <button class="danger" on:click={() => remove(m)}>Delete</button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
