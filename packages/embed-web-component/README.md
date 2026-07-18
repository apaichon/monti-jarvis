# @monti/embed-web-component

Custom element `<monti-embed>` for plain HTML, Angular, and other hosts.

## Install

```bash
npm install @monti/embed-web-component @monti/embed-core
```

## Usage

```html
<script type="module">
  import "@monti/embed-web-component";
</script>

<monti-embed
  embed-key="emb_YOUR_KEY"
  api-base="http://localhost:8091"
  position="bottom-right"
></monti-embed>

<script>
  const el = document.querySelector("monti-embed");
  el.addEventListener("monti-open", () => console.log("open"));
  el.addEventListener("monti-error", (e) => console.error(e.detail));
  // el.open(); el.close();
</script>
```

### Attributes

| Attribute | Description |
| --- | --- |
| `embed-key` | Required |
| `api-base` | Required Monti origin |
| `parent-origin` | Optional host origin |
| `position` | Floating corner |
| `agent-id` | Optional agent |
| `theme` / `locale` | Optional query hints |
| `open` | Start open (boolean attr) |
| `inline` | Inline panel in place of floating launcher |
| `skip-resolve` | Skip pre-resolve |

### Events

`monti-open`, `monti-close`, `monti-ready`, `monti-error`, `monti-destroy`
