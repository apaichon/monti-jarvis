/**
 * Monti Jarvis embed loader (SPRINT-014).
 * Usage:
 *   <script src="{host}/embed/monti-embed.js"
 *     data-embed-key="emb_…"
 *     data-position="bottom-right"
 *     async></script>
 */
(function () {
  'use strict';
  if (window.__montiEmbedLoaded) return;
  window.__montiEmbedLoaded = true;

  var script =
    document.currentScript ||
    (function () {
      var list = document.getElementsByTagName('script');
      return list[list.length - 1];
    })();

  var key = (script && script.getAttribute('data-embed-key')) || '';
  var position = (script && script.getAttribute('data-position')) || 'bottom-right';
  var base =
    (script && script.getAttribute('data-base')) ||
    (script && script.src ? script.src.replace(/\/embed\/monti-embed\.js.*$/i, '') : '') ||
    window.location.origin;

  if (!key) {
    console.warn('[monti-embed] missing data-embed-key');
    return;
  }

  var open = false;
  var z = 2147483000;
  var root = document.createElement('div');
  root.id = 'monti-embed-root';
  root.setAttribute('data-monti-embed', '1');
  root.style.cssText =
    'all:initial;position:fixed;z-index:' + z + ';font-family:system-ui,sans-serif;';

  var positions = {
    'bottom-right': { right: '20px', bottom: '20px', left: 'auto', top: 'auto' },
    'bottom-left': { left: '20px', bottom: '20px', right: 'auto', top: 'auto' },
    'top-right': { right: '20px', top: '20px', left: 'auto', bottom: 'auto' },
    'top-left': { left: '20px', top: '20px', right: 'auto', bottom: 'auto' }
  };
  var pos = positions[position] || positions['bottom-right'];

  // Panel holds iframe + close control (close is NOT over the Send button).
  var panel = document.createElement('div');
  panel.style.cssText =
    'display:none;position:fixed;' +
    'right:' +
    (pos.right || 'auto') +
    ';bottom:' +
    (pos.bottom || 'auto') +
    ';left:' +
    (pos.left || 'auto') +
    ';top:' +
    (pos.top || 'auto') +
    ';width:min(400px,calc(100vw - 24px));height:min(680px,calc(100vh - 40px));' +
    'border-radius:16px;overflow:hidden;box-shadow:0 16px 48px rgba(0,0,0,.35);' +
    'border:1px solid rgba(0,183,255,.35);background:#05101f;z-index:' +
    (z + 1) +
    ';';

  var iframe = document.createElement('iframe');
  iframe.title = 'Monti AI Assistant';
  // Permissions Policy: allow mic + autoplay inside cross-origin iframe.
  // Set BEFORE src. "*" allows any origin loaded in this frame (needed for Monti host).
  var allowFeatures =
    'microphone *; autoplay *; camera *; clipboard-write *; display-capture *';
  iframe.setAttribute('allow', allowFeatures);
  iframe.allow = allowFeatures;
  // Legacy Feature Policy attribute (older Chromium)
  iframe.setAttribute(
    'allowfullscreen',
    'true'
  );
  iframe.setAttribute('referrerpolicy', 'strict-origin-when-cross-origin');
  iframe.style.cssText = 'width:100%;height:100%;border:0;display:block;background:#05101f;';
  // Host origin for allowlist (iframe document origin is Monti, not the host site).
  var parentOrigin = window.location.origin || '';
  // Prefer secure base when parent is secure but data-base/script is http custom host —
  // callers should still use https or localhost for Monti (see docs).
  iframe.src =
    base +
    '/embed?key=' +
    encodeURIComponent(key) +
    (parentOrigin ? '&parent_origin=' + encodeURIComponent(parentOrigin) : '');
  panel.appendChild(iframe);

  // Warn integrators early if Monti host cannot expose mediaDevices (non-secure context).
  try {
    var montiUrl = new URL(base, window.location.href);
    var montiHost = (montiUrl.hostname || '').toLowerCase();
    var montiSecure =
      montiUrl.protocol === 'https:' ||
      montiHost === 'localhost' ||
      montiHost === '127.0.0.1' ||
      montiHost.endsWith('.localhost');
    if (!montiSecure) {
      console.warn(
        '[monti-embed] Voice/mic will fail: Monti host is not a secure context (' +
          montiUrl.origin +
          '). Use https:// or http://localhost:PORT for the script src / data-base.'
      );
    }
  } catch (e) {
    /* ignore */
  }

  // Close sits on top-right of the panel — never over the composer Send button.
  var closeBtn = document.createElement('button');
  closeBtn.type = 'button';
  closeBtn.setAttribute('aria-label', 'Close Monti chat');
  closeBtn.textContent = '✕';
  closeBtn.style.cssText =
    'position:absolute;top:10px;right:10px;z-index:3;width:32px;height:32px;border-radius:50%;' +
    'border:1px solid rgba(0,183,255,.4);cursor:pointer;background:rgba(8,20,36,.92);' +
    'color:#f7fbff;font-size:14px;line-height:1;box-shadow:0 4px 12px rgba(0,0,0,.35);' +
    'display:none;padding:0;';
  panel.style.position = 'fixed';
  panel.appendChild(closeBtn);

  // Floating launcher when panel is closed.
  var openBtn = document.createElement('button');
  openBtn.type = 'button';
  openBtn.setAttribute('aria-label', 'Open Monti chat');
  openBtn.textContent = '💬';
  openBtn.style.cssText =
    'position:fixed;right:' +
    (pos.right || 'auto') +
    ';bottom:' +
    (pos.bottom || 'auto') +
    ';left:' +
    (pos.left || 'auto') +
    ';top:' +
    (pos.top || 'auto') +
    ';width:56px;height:56px;border-radius:50%;border:0;cursor:pointer;' +
    'background:linear-gradient(135deg,#0084ff,#00b7ff);color:#fff;font-size:22px;' +
    'box-shadow:0 8px 24px rgba(0,120,255,.45);z-index:' +
    (z + 2) +
    ';';

  function setOpen(next) {
    open = next;
    panel.style.display = open ? 'block' : 'none';
    closeBtn.style.display = open ? 'block' : 'none';
    openBtn.style.display = open ? 'none' : 'block';
  }

  openBtn.addEventListener('click', function () {
    setOpen(true);
  });
  closeBtn.addEventListener('click', function () {
    setOpen(false);
  });

  root.appendChild(panel);
  root.appendChild(openBtn);
  document.body.appendChild(root);
})();
