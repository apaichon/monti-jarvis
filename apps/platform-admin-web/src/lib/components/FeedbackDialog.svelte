<script lang="ts">
  import Dialog from '$lib/components/ui/Dialog.svelte';
  import { feedback } from '$lib/feedback.svelte';

  const icons: Record<string, string> = {
    success: '✓',
    error: '!',
    info: 'i'
  };
</script>

<Dialog open={feedback.open} onClose={() => feedback.close()}>
  <div class="dialog-content" role="alertdialog" aria-labelledby="feedback-title" aria-describedby="feedback-desc">
    <div class={`dialog-icon dialog-icon-${feedback.variant}`} aria-hidden="true">
      {icons[feedback.variant]}
    </div>
    <div class="dialog-header">
      <h2 id="feedback-title" class="dialog-title">{feedback.title}</h2>
      <p id="feedback-desc" class="dialog-description">{feedback.message}</p>
    </div>
    <div class="dialog-footer">
      <button type="button" class="btn dialog-btn" onclick={() => feedback.close()}>OK</button>
    </div>
  </div>
</Dialog>