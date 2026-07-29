import {
  getCurrentCommercialPlan,
  type CurrentCommercialPlan
} from '$lib/api/billing';

class CurrentPlanStore {
  data = $state<CurrentCommercialPlan | null>(null);
  loading = $state(false);
  error = $state('');
  private pending: Promise<CurrentCommercialPlan> | null = null;

  async load(force = false): Promise<CurrentCommercialPlan> {
    if (this.data && !force) return this.data;
    if (this.pending && !force) return this.pending;

    this.loading = true;
    this.error = '';
    this.pending = getCurrentCommercialPlan();
    try {
      this.data = await this.pending;
      return this.data;
    } catch (error) {
      this.error = error instanceof Error ? error.message : 'Unable to load current plan';
      throw error;
    } finally {
      this.loading = false;
      this.pending = null;
    }
  }

  clear() {
    this.data = null;
    this.error = '';
    this.pending = null;
  }
}

export const currentPlan = new CurrentPlanStore();
