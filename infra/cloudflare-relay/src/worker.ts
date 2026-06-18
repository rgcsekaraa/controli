export interface Env {
  SESSIONS: DurableObjectNamespace;
  INVITES: DurableObjectNamespace;
  INVITES_KV: KVNamespace;
}

type Side = "host" | "client";
type RelayPayload = string | ArrayBuffer;
interface InvitePayload {
  kind: "controli-relay-token";
  version: number;
  code: string;
  session_id: string;
  name: string;
  room?: string;
  mode?: string;
  transport?: string;
  tunnel_url?: string;
  relay_url: string;
  secret: string;
  expires_at: string;
  invite_expires_at: string;
}

const SIDES = new Set(["host", "client"]);
const FINAL_CLOSE_CODE = 1000;
const FINAL_CLOSE_REASON = "controli-final-close";
const MAX_PENDING_MESSAGES = 2048;
const MAX_PENDING_BYTES = 16 * 1024 * 1024;
const CONTROL_PREFIX = "\x00CONTROLI:";
const CONTROL_GUEST_CONNECTED = "guest_connected";
const CONTROL_GUEST_DISCONNECTED = "guest_disconnected";

function jsonResponse(status: number, payload: object): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { "content-type": "application/json" },
  });
}

function getParam(url: URL, name: string): string | null {
  const value = url.searchParams.get(name);
  return value && value.trim() ? value : null;
}

function isSide(value: string | null): value is Side {
  return value !== null && SIDES.has(value);
}

function normalizeCode(value: string): string | null {
  const code = value.replace(/\D/g, "");
  return /^\d{7}$/.test(code) ? code : null;
}

function textField(payload: Record<string, unknown>, name: string): string | null {
  const value = payload[name];
  return typeof value === "string" && value.trim() ? value : null;
}

function buildInvite(payload: Record<string, unknown> | null): InvitePayload | null {
  if (!payload) {
    return null;
  }
  const code = textField(payload, "code");
  const sessionId = textField(payload, "session_id");
  const name = textField(payload, "name") ?? "guest";
  const room = textField(payload, "room") ?? undefined;
  const mode = textField(payload, "mode") ?? undefined;
  const transport = textField(payload, "transport") ?? undefined;
  const tunnelUrl = textField(payload, "tunnel_url") ?? undefined;
  const relayUrl = textField(payload, "relay_url");
  const secret = textField(payload, "secret");
  const expiresAt = textField(payload, "expires_at");
  const inviteExpiresAt = textField(payload, "invite_expires_at");
  if (!code || !sessionId || !relayUrl || !secret || !expiresAt || !inviteExpiresAt) {
    return null;
  }
  return {
    kind: "controli-relay-token",
    version: 1,
    code,
    session_id: sessionId,
    name,
    room,
    mode,
    transport,
    tunnel_url: tunnelUrl,
    relay_url: relayUrl,
    secret,
    expires_at: expiresAt,
    invite_expires_at: inviteExpiresAt,
  };
}

function isExpired(value: string): boolean {
  const time = Date.parse(value);
  return Number.isNaN(time) || time < Date.now();
}

function inviteKey(code: string): string {
  return `invite:${code}`;
}

function inviteTTL(value: string): number {
  const seconds = Math.floor((Date.parse(value) - Date.now()) / 1000);
  return Math.max(0, seconds);
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === "/health") {
      return jsonResponse(200, {
        ok: true,
        service: "controli-relay",
        websocket_path: "/v1/ws",
        invite_store: "kv",
        relay_clients: "single-active",
        max_pending_messages: MAX_PENDING_MESSAGES,
        max_pending_bytes: MAX_PENDING_BYTES,
      });
    }

    if (url.pathname === "/v1/invite/register" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const code = typeof payload?.code === "string" ? normalizeCode(payload.code) : null;
      if (!code) {
        return jsonResponse(400, { error: "valid 7-digit code is required" });
      }
      const invite = buildInvite({ ...payload, code });
      if (!invite) {
        return jsonResponse(400, { error: "invalid invite payload" });
      }
      const existing = await env.INVITES_KV.get<InvitePayload>(inviteKey(code), "json");
      if (existing && !isExpired(existing.invite_expires_at)) {
        return jsonResponse(409, { error: "short code collision; try again" });
      }
      const ttl = inviteTTL(invite.invite_expires_at);
      if (ttl <= 0) {
        return jsonResponse(400, { error: "invite is already expired" });
      }
      await env.INVITES_KV.put(inviteKey(code), JSON.stringify(invite), { expirationTtl: ttl });
      return jsonResponse(200, { ok: true });
    }

    if (url.pathname === "/v1/invite/claim" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const code = typeof payload?.code === "string" ? normalizeCode(payload.code) : null;
      if (!code) {
        return jsonResponse(400, { error: "valid 7-digit code is required" });
      }
      const invite = await env.INVITES_KV.get<InvitePayload>(inviteKey(code), "json");
      if (!invite) {
        return jsonResponse(404, { error: "unknown short code" });
      }
      if (isExpired(invite.invite_expires_at)) {
        await env.INVITES_KV.delete(inviteKey(code));
        return jsonResponse(410, { error: "short code expired" });
      }
      return jsonResponse(200, invite);
    }

    if (url.pathname === "/v1/close" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const sessionId = typeof payload?.session_id === "string" ? payload.session_id : null;
      if (!sessionId) {
        return jsonResponse(400, { error: "session_id is required" });
      }
      const id = env.SESSIONS.idFromName(sessionId);
      const object = env.SESSIONS.get(id);
      return object.fetch(
        new Request(request.url, {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify(payload),
        }),
      );
    }

    if (url.pathname === "/v1/session/status" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const sessionId = typeof payload?.session_id === "string" ? payload.session_id : null;
      if (!sessionId) {
        return jsonResponse(400, { error: "session_id is required" });
      }
      const object = env.SESSIONS.get(env.SESSIONS.idFromName(sessionId));
      return object.fetch(
        new Request(request.url, {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify(payload),
        }),
      );
    }

    if (url.pathname !== "/v1/ws") {
      return jsonResponse(404, { error: "not found" });
    }

    if (request.headers.get("Upgrade")?.toLowerCase() !== "websocket") {
      return jsonResponse(426, { error: "expected websocket upgrade" });
    }

    const sessionId = getParam(url, "session_id");
    const secret = getParam(url, "secret");
    const side = getParam(url, "side");
    if (!sessionId || !secret || !isSide(side)) {
      return jsonResponse(400, { error: "session_id, secret, and side are required" });
    }

    const id = env.SESSIONS.idFromName(sessionId);
    const object = env.SESSIONS.get(id);
    return object.fetch(request);
  },
};

export class InviteCode implements DurableObject {
  private readonly state: DurableObjectState;

  constructor(state: DurableObjectState) {
    this.state = state;
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    const payload = await request.json<Record<string, unknown>>().catch(() => null);
    const code = typeof payload?.code === "string" ? normalizeCode(payload.code) : null;
    if (!code) {
      return jsonResponse(400, { error: "valid 7-digit code is required" });
    }

    if (url.pathname === "/v1/invite/register" && request.method === "POST") {
      const invite = buildInvite(payload);
      if (!invite) {
        return jsonResponse(400, { error: "invalid invite payload" });
      }
      const existing = await this.state.storage.get<InvitePayload>("invite");
      if (existing && !isExpired(existing.invite_expires_at)) {
        return jsonResponse(409, { error: "short code collision; try again" });
      }
      await this.state.storage.put("invite", invite);
      return jsonResponse(200, { ok: true });
    }

    if (url.pathname === "/v1/invite/claim" && request.method === "POST") {
      const invite = await this.state.storage.get<InvitePayload>("invite");
      if (!invite) {
        return jsonResponse(404, { error: "unknown short code" });
      }
      if (isExpired(invite.invite_expires_at)) {
        await this.state.storage.delete("invite");
        return jsonResponse(410, { error: "short code expired" });
      }
      return jsonResponse(200, invite);
    }

    return jsonResponse(404, { error: "not found" });
  }
}

export class RelaySession implements DurableObject {
  private readonly state: DurableObjectState;
  private sockets: Map<string, WebSocket> = new Map();
  private pending: Map<Side, RelayPayload[]> = new Map([
    ["host", []],
    ["client", []],
  ]);
  private pendingBytes: Map<Side, number> = new Map([
    ["host", 0],
    ["client", 0],
  ]);
  private secretHash: string | null = null;
  private connectedAt: Map<string, string> = new Map();
  private lastActivityAt: string | null = null;
  private droppedMessages = 0;

  constructor(state: DurableObjectState) {
    this.state = state;
    this.state.blockConcurrencyWhile(async () => {
      this.secretHash = (await this.state.storage.get<string>("secretHash")) ?? null;
    });
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    if (url.pathname === "/v1/session/status" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const secret = typeof payload?.secret === "string" ? payload.secret : null;
      if (!secret) {
        return jsonResponse(400, { error: "secret is required" });
      }
      if (!(await this.authorize(secret))) {
        return jsonResponse(403, { error: "invalid session secret" });
      }
      return jsonResponse(200, this.status());
    }

    if (url.pathname === "/v1/close" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const secret = typeof payload?.secret === "string" ? payload.secret : null;
      const side = typeof payload?.side === "string" ? payload.side : null;
      const clientId = typeof payload?.client_id === "string" ? payload.client_id : null;
      if (!secret || !isSide(side)) {
        return jsonResponse(400, { error: "secret and side are required" });
      }
      if (!(await this.authorize(secret))) {
        return jsonResponse(403, { error: "invalid session secret" });
      }
      this.finalClose(side, clientId);
      return jsonResponse(200, { ok: true });
    }

    const sessionId = getParam(url, "session_id");
    const secret = getParam(url, "secret");
    const side = getParam(url, "side");
    const clientId = getParam(url, "client_id");
    if (!sessionId || !secret || !isSide(side)) {
      return jsonResponse(400, { error: "session_id, secret, and side are required" });
    }

    if (!(await this.authorize(secret))) {
      return jsonResponse(403, { error: "invalid session secret" });
    }

    if (side === "client" && this.hasActiveClient()) {
      return jsonResponse(409, { error: "session already has an active guest" });
    }

    const pair = new WebSocketPair();
    const [client, server] = Object.values(pair);
    this.attach(side, server, clientId);
    return new Response(null, { status: 101, webSocket: client });
  }

  private async authorize(secret: string): Promise<boolean> {
    const digest = await sha256(secret);
    if (this.secretHash === null) {
      this.secretHash = digest;
      await this.state.storage.put("secretHash", digest);
      await this.state.storage.put("createdAt", new Date().toISOString());
      return true;
    }
    return timingSafeEqual(this.secretHash, digest);
  }

  private attach(side: Side, socket: WebSocket, clientId: string | null): void {
    const key = socketKey(side, clientId);
    const old = side === "host" ? this.sockets.get(key) : null;
    if (old) {
      old.close(1012, "replaced");
    }

    socket.accept();
    this.sockets.set(key, socket);
    this.connectedAt.set(key, new Date().toISOString());
    if (side === "client") {
      this.notifyHost(CONTROL_GUEST_CONNECTED);
    }
    this.flush(side, key);

    socket.addEventListener("message", (event) => {
      this.forward(side, event.data).catch(() => {
        socket.close(1003, "unsupported message");
      });
    });

    socket.addEventListener("close", (event) => this.detach(side, key, event.code, event.reason));
    socket.addEventListener("error", () => this.detach(side, key, 1011, "socket error"));
  }

  private detach(side: Side, key: string, code: number, reason: string): void {
    const socket = this.sockets.get(key);
    this.sockets.delete(key);
    this.connectedAt.delete(key);
    const finalClose = code === FINAL_CLOSE_CODE && reason === FINAL_CLOSE_REASON;
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.close(1000, "closed");
    }
    if (!finalClose) {
      if (side === "client" && socket) {
        this.notifyHost(CONTROL_GUEST_DISCONNECTED);
      }
      return;
    }
    if (side === "host") {
      this.closeClients("host disconnected");
      return;
    }
    if (side === "client" && socket) {
      this.notifyHost(CONTROL_GUEST_DISCONNECTED);
    }
  }

  private async forward(from: Side, data: unknown): Promise<void> {
    const payload = await normalizePayload(data);
    this.lastActivityAt = new Date().toISOString();
    if (from === "host") {
      this.deliverClients(payload);
      return;
    }
    this.deliverHost(payload);
  }

  private deliverHost(payload: RelayPayload): void {
    const socket = this.sockets.get("host");
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(payload);
      return;
    }
    this.queuePending("host", payload);
  }

  private deliverClients(payload: RelayPayload): void {
    const clients = this.clientSockets();
    if (clients.length > 0) {
      for (const socket of clients) {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(payload);
        }
      }
      return;
    }
    this.queuePending("client", payload);
  }

  private queuePending(to: Side, payload: RelayPayload): void {
    const queue = this.pending.get(to) ?? [];
    const bytes = payloadBytes(payload);
    queue.push(payload);
    this.pendingBytes.set(to, (this.pendingBytes.get(to) ?? 0) + bytes);
    while (queue.length > MAX_PENDING_MESSAGES || (this.pendingBytes.get(to) ?? 0) > MAX_PENDING_BYTES) {
      const dropped = queue.shift();
      if (dropped !== undefined) {
        this.pendingBytes.set(to, Math.max(0, (this.pendingBytes.get(to) ?? 0) - payloadBytes(dropped)));
        this.droppedMessages += 1;
      }
    }
    this.pending.set(to, queue);
  }

  private flush(side: Side, key: string): void {
    const socket = this.sockets.get(key);
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    const queue = this.pending.get(side) ?? [];
    while (queue.length > 0 && socket.readyState === WebSocket.OPEN) {
      const payload = queue.shift();
      if (payload !== undefined) {
        socket.send(payload);
        this.pendingBytes.set(side, Math.max(0, (this.pendingBytes.get(side) ?? 0) - payloadBytes(payload)));
      }
    }
  }

  private finalClose(side: Side, clientId: string | null): void {
    if (side === "host") {
      const host = this.sockets.get("host");
      this.sockets.delete("host");
      this.connectedAt.delete("host");
      this.pending.set("host", []);
      this.pending.set("client", []);
      this.pendingBytes.set("host", 0);
      this.pendingBytes.set("client", 0);
      if (host && host.readyState === WebSocket.OPEN) {
        host.close(FINAL_CLOSE_CODE, "host disconnected");
      }
      this.closeClients("host disconnected");
      return;
    }
    const key = socketKey("client", clientId);
    const socket = this.sockets.get(key);
    this.sockets.delete(key);
    this.connectedAt.delete(key);
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.close(FINAL_CLOSE_CODE, "client disconnected");
    }
    this.notifyHost(CONTROL_GUEST_DISCONNECTED);
  }

  private status(): object {
    return {
      ok: true,
      connected: {
        host: this.sockets.has("host"),
        clients: this.clientSockets().length,
      },
      connected_at: {
        host: this.connectedAt.get("host") ?? null,
      },
      pending_messages: {
        host: this.pending.get("host")?.length ?? 0,
        client: this.pending.get("client")?.length ?? 0,
      },
      pending_bytes: {
        host: this.pendingBytes.get("host") ?? 0,
        client: this.pendingBytes.get("client") ?? 0,
      },
      last_activity_at: this.lastActivityAt,
      dropped_messages: this.droppedMessages,
      limits: {
        max_pending_messages: MAX_PENDING_MESSAGES,
        max_pending_bytes: MAX_PENDING_BYTES,
      },
    };
  }

  private clientSockets(): WebSocket[] {
    const sockets: WebSocket[] = [];
    for (const [key, socket] of this.sockets) {
      if (isClientKey(key)) {
        sockets.push(socket);
      }
    }
    return sockets;
  }

  private hasActiveClient(): boolean {
    return this.clientSockets().some((socket) => socket.readyState === WebSocket.OPEN);
  }

  private notifyHost(type: string): void {
    this.deliverHost(`${CONTROL_PREFIX}${JSON.stringify({ type })}`);
  }

  private closeClients(reason: string): void {
    for (const [key, socket] of this.sockets) {
      if (!isClientKey(key)) {
        continue;
      }
      this.sockets.delete(key);
      this.connectedAt.delete(key);
      if (socket.readyState === WebSocket.OPEN) {
        socket.close(FINAL_CLOSE_CODE, reason);
      }
    }
  }
}

function socketKey(side: Side, clientId: string | null): string {
  if (side === "host") {
    return "host";
  }
  const trimmed = clientId?.trim();
  return trimmed ? `client:${trimmed}` : "client";
}

function isClientKey(key: string): boolean {
  return key === "client" || key.startsWith("client:");
}

function payloadBytes(payload: RelayPayload): number {
  if (typeof payload === "string") {
    return new TextEncoder().encode(payload).byteLength;
  }
  return payload.byteLength;
}

async function sha256(value: string): Promise<string> {
  const input = new TextEncoder().encode(value);
  const digest = await crypto.subtle.digest("SHA-256", input);
  return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

function timingSafeEqual(left: string, right: string): boolean {
  if (left.length !== right.length) {
    return false;
  }
  let diff = 0;
  for (let index = 0; index < left.length; index += 1) {
    diff |= left.charCodeAt(index) ^ right.charCodeAt(index);
  }
  return diff === 0;
}

async function normalizePayload(data: unknown): Promise<RelayPayload> {
  if (typeof data === "string") {
    return data;
  }
  if (data instanceof ArrayBuffer) {
    return data;
  }
  if (ArrayBuffer.isView(data)) {
    const view = new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
    const copy = new Uint8Array(view.byteLength);
    copy.set(view);
    return copy.buffer;
  }
  if (data instanceof Blob) {
    return data.arrayBuffer();
  }
  return String(data);
}
