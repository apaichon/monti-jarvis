<script lang="ts">
  import type { AvatarVoice } from '$lib/api/avatars';
  import { defaultVoiceRow } from '$lib/api/avatars';

  let { voices = $bindable<AvatarVoice[]>([]) } = $props();

  function addRow() {
    const nextPriority = voices.length ? Math.max(...voices.map((v) => v.priority)) + 1 : 1;
    voices = [...voices, defaultVoiceRow(nextPriority)];
  }

  function removeRow(index: number) {
    voices = voices.filter((_, i) => i !== index);
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
  <h2 style="margin:0;font-size:16px">Voice profiles</h2>
  <button class="btn ghost" type="button" onclick={addRow}>+ Add row</button>
</div>

{#if voices.length === 0}
  <p style="color:var(--muted);font-size:13px;margin:0 0 12px">At least one voice profile is required.</p>
{:else}
  <table class="table">
    <thead>
      <tr>
        <th>priority</th>
        <th>provider</th>
        <th>voice_id</th>
        <th>voice</th>
        <th>status</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#each voices as voice, index (index)}
        <tr>
          <td>
            <input
              type="number"
              min="1"
              bind:value={voice.priority}
              style="width:64px;padding:6px 8px"
            />
          </td>
          <td>
            <input bind:value={voice.voice_provider_id} placeholder="voice-gemini-live" required />
          </td>
          <td>
            <input bind:value={voice.voice_id} placeholder="gemini-2.5-flash-native-audio-latest" required />
          </td>
          <td>
            <input bind:value={voice.voice} placeholder="Aoede" required />
          </td>
          <td>
            <select bind:value={voice.status}>
              <option value="active">active</option>
              <option value="disabled">disabled</option>
            </select>
          </td>
          <td>
            {#if voices.length > 1}
              <button
                class="link"
                type="button"
                style="background:none;border:none;padding:0;color:var(--danger)"
                onclick={() => removeRow(index)}
              >
                Remove
              </button>
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}