export type FeedbackVariant = 'success' | 'error' | 'info';

const titles: Record<FeedbackVariant, string> = {
  success: 'Success',
  error: 'Error',
  info: 'Notice'
};

class FeedbackStore {
  open = $state(false);
  variant = $state<FeedbackVariant>('info');
  title = $state('');
  message = $state('');

  show(variant: FeedbackVariant, message: string, title?: string) {
    this.variant = variant;
    this.message = message;
    this.title = title ?? titles[variant];
    this.open = true;
  }

  success(message: string, title?: string) {
    this.show('success', message, title);
  }

  error(message: string, title?: string) {
    this.show('error', message, title);
  }

  info(message: string, title?: string) {
    this.show('info', message, title);
  }

  close() {
    this.open = false;
  }
}

export const feedback = new FeedbackStore();