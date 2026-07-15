# Monti Mobile SDK

The SDK is a dependency-free TypeScript core for iOS, Android, React Native, and Flutter host adapters. Hosts provide token storage, `fetch`, and a WebSocket factory; authentication, tenant policy, avatar assignment, quota, call lifecycle, transcript, and review behavior remain server-owned.

```ts
const client = new MontiMobileClient({
  baseUrl: "https://support.example.com",
  tenantId: "tenant-slug", // omit in authenticated production flows
  tokenStore,
  websocket: (url) => new WebSocket(url),
});

const bootstrap = await client.getBootstrap();
const call = await client.createCall({ avatarId: bootstrap.default_avatar_id });
const handle = await client.connectCall(call.call_id);
handle.onEvent((event) => console.log(event));
```

Use the host platform's secure keychain/keystore for `TokenStore`. Never put Gemini, LiveKit, MinIO, or database credentials in a mobile application.
