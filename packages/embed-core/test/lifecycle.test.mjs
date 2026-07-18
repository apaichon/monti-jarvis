import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  MontiEmbed,
  buildEmbedIframeUrl,
  buildResolveUrl,
  normalizeApiBase,
} from "../dist/index.js";

describe("normalizeApiBase", () => {
  it("strips trailing slashes", () => {
    assert.equal(normalizeApiBase("http://localhost:8091/"), "http://localhost:8091");
    assert.equal(normalizeApiBase("https://monti.example.com///"), "https://monti.example.com");
  });
});

describe("buildEmbedIframeUrl", () => {
  it("builds iframe URL with key and parent_origin", () => {
    const url = buildEmbedIframeUrl({
      apiBase: "http://localhost:8091/",
      embedKey: "emb_abc",
      parentOrigin: "http://localhost:5173",
      agentId: "ava",
      theme: "dark",
      locale: "en",
    });
    const u = new URL(url);
    assert.equal(u.origin + u.pathname, "http://localhost:8091/embed");
    assert.equal(u.searchParams.get("key"), "emb_abc");
    assert.equal(u.searchParams.get("parent_origin"), "http://localhost:5173");
    assert.equal(u.searchParams.get("agent"), "ava");
    assert.equal(u.searchParams.get("theme"), "dark");
    assert.equal(u.searchParams.get("locale"), "en");
  });
});

describe("buildResolveUrl", () => {
  it("encodes key in path", () => {
    const url = buildResolveUrl("http://localhost:8091", "emb_x/y", "https://shop.example");
    assert.match(url, /\/api\/public\/embed\/emb_x%2Fy/);
    assert.match(url, /parent_origin=https%3A%2F%2Fshop\.example/);
  });
});

function stubDom() {
  const body = {
    children: [],
    appendChild(el) {
      this.children.push(el);
      return el;
    },
  };
  const doc = {
    createElement(tag) {
      const el = {
        style: { cssText: "" },
        tagName: tag.toUpperCase(),
        children: [],
        setAttribute() {},
        addEventListener() {},
        removeEventListener() {},
        appendChild(child) {
          this.children.push(child);
          return child;
        },
        remove() {
          const idx = body.children.indexOf(el);
          if (idx >= 0) body.children.splice(idx, 1);
        },
        innerHTML: "",
      };
      if (tag === "iframe") {
        el.allow = "";
        el.src = "";
        el.title = "";
      }
      return el;
    },
    body,
  };
  globalThis.document = doc;
  globalThis.window = {
    location: { origin: "http://localhost:5173", href: "http://localhost:5173/" },
  };
  return body;
}

describe("MontiEmbed lifecycle (mock DOM)", () => {
  it("open/close/destroy without leaking listeners", async () => {
    const body = stubDom();
    const mockFetch = async () =>
      new Response(
        JSON.stringify({
          tenant_id: "demo",
          slug: "demo",
          name: "Demo",
          embed_key: "emb_test",
          enabled: true,
          default_agent_id: "ava",
          agents: [],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );

    const events = [];
    const embed = new MontiEmbed({
      embedKey: "emb_test",
      apiBase: "http://localhost:8091",
      parentOrigin: "http://localhost:5173",
      fetch: mockFetch,
      onOpen: () => events.push("open"),
      onClose: () => events.push("close"),
      onReady: () => events.push("ready"),
      onDestroy: () => events.push("destroy"),
    });

    await embed.mount();
    assert.equal(events.includes("ready"), true);
    assert.equal(embed.isOpen, false);

    embed.open();
    assert.equal(embed.isOpen, true);
    assert.equal(events.filter((e) => e === "open").length, 1);

    embed.close();
    assert.equal(embed.isOpen, false);
    assert.equal(events.filter((e) => e === "close").length, 1);

    embed.destroy();
    assert.equal(embed.isDestroyed, true);
    assert.equal(events.includes("destroy"), true);
    assert.equal(body.children.length, 0);

    embed.destroy();
    assert.equal(events.filter((e) => e === "destroy").length, 1);
  });

  it("surfaces clear error for bad embed key", async () => {
    stubDom();
    const mockFetch = async () =>
      new Response(JSON.stringify({ code: "embed_not_found", error: "Unknown embed key" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });

    let errCode = "";
    const embed = new MontiEmbed({
      embedKey: "emb_bad",
      apiBase: "http://localhost:8091",
      fetch: mockFetch,
      onError: (e) => {
        errCode = e.code;
      },
    });
    await embed.mount();
    assert.equal(errCode, "embed_not_found");
    assert.equal(embed.error?.code, "embed_not_found");
    embed.destroy();
  });
});
