# @monti/embed-react

React component + hooks for the Monti Jarvis web embed.

## Install

```bash
npm install @monti/embed-react @monti/embed-core
```

## Usage

```tsx
import { MontiEmbedReact } from "@monti/embed-react";

export function SupportWidget() {
  return (
    <MontiEmbedReact
      embedKey="emb_YOUR_KEY"
      apiBase="http://localhost:8091"
      position="bottom-right"
      onError={(e) => console.error(e.code, e.message)}
    />
  );
}
```

### Hook

```tsx
import { useMontiEmbed } from "@monti/embed-react";

function Custom() {
  const { open, close, error } = useMontiEmbed({
    embedKey: "emb_YOUR_KEY",
    apiBase: "http://localhost:8091",
  });
  return (
    <>
      <button onClick={open}>Chat</button>
      {error && <p>{error.message}</p>}
    </>
  );
}
```
