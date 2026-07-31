// @ts-nocheck -- executed by Node's built-in test runner; app has no Node type package.
import assert from 'node:assert/strict';
import test from 'node:test';

import { normalizeLeadListResponse } from '../src/lib/api/lead-list-contract.js';

test('keeps API items and total for a non-empty Book Demo response', () => {
  const lead = {
    id: 'lead_1',
    kind: 'book_demo',
    status: 'new',
    email: 'demo@example.com'
  };

  const result = normalizeLeadListResponse({
    items: [lead],
    total: 1,
    limit: 50,
    offset: 0
  });

  assert.deepEqual(result.items, [lead]);
  assert.equal(result.total, 1);
  assert.equal(result.limit, 50);
  assert.equal(result.offset, 0);
});

test('uses rendered item count when the server omits total', () => {
  const result = normalizeLeadListResponse({
    items: [{ id: 'lead_1' }, { id: 'lead_2' }]
  });

  assert.equal(result.items.length, 2);
  assert.equal(result.total, 2);
});

test('normalizes a null items collection to a true empty result', () => {
  const result = normalizeLeadListResponse({ items: null, total: 0 });

  assert.deepEqual(result.items, []);
  assert.equal(result.total, 0);
});
