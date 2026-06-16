export interface Env {
  SESSIONS: DurableObjectNamespace;
  INVITES: DurableObjectNamespace;
}

type Side = "host" | "client";
type RelayPayload = string | ArrayBuffer;
interface InvitePayload {
  kind: "controli-relay-token";
  version: number;
  code: string;
  session_id: string;
  name: string;
  relay_url: string;
  secret: string;
  expires_at: string;
  invite_expires_at: string;
}

const SIDES = new Set(["host", "client"]);
const FINAL_CLOSE_CODE = 1000;
const FINAL_CLOSE_REASON = "controli-final-close";
const MAX_PENDING_MESSAGES = 2048;

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

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === "/health") {
      return jsonResponse(200, { ok: true });
    }

    if (url.pathname === "/v1/invite/register" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const code = typeof payload?.code === "string" ? normalizeCode(payload.code) : null;
      if (!code) {
        return jsonResponse(400, { error: "valid 7-digit code is required" });
      }
      const object = env.INVITES.get(env.INVITES.idFromName(code));
      return object.fetch(
        new Request(request.url, {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify({ ...payload, code }),
        }),
      );
    }

    if (url.pathname === "/v1/invite/claim" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const code = typeof payload?.code === "string" ? normalizeCode(payload.code) : null;
      if (!code) {
        return jsonResponse(400, { error: "valid 7-digit code is required" });
      }
      const object = env.INVITES.get(env.INVITES.idFromName(code));
      return object.fetch(
        new Request(request.url, {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify({ code }),
        }),
      );
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
  private sockets: Map<Side, WebSocket> = new Map();
  private pending: Map<Side, RelayPayload[]> = new Map([
    ["host", []],
    ["client", []],
  ]);
  private secretHash: string | null = null;

  constructor(state: DurableObjectState) {
    this.state = state;
    this.state.blockConcurrencyWhile(async () => {
      this.secretHash = (await this.state.storage.get<string>("secretHash")) ?? null;
    });
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    if (url.pathname === "/v1/close" && request.method === "POST") {
      const payload = await request.json<Record<string, unknown>>().catch(() => null);
      const secret = typeof payload?.secret === "string" ? payload.secret : null;
      const side = typeof payload?.side === "string" ? payload.side : null;
      if (!secret || !isSide(side)) {
        return jsonResponse(400, { error: "secret and side are required" });
      }
      if (!(await this.authorize(secret))) {
        return jsonResponse(403, { error: "invalid session secret" });
      }
      this.finalClose(side);
      return jsonResponse(200, { ok: true });
    }

    const sessionId = getParam(url, "session_id");
    const secret = getParam(url, "secret");
    const side = getParam(url, "side");
    if (!sessionId || !secret || !isSide(side)) {
      return jsonResponse(400, { error: "session_id, secret, and side are required" });
    }

    if (!(await this.authorize(secret))) {
      return jsonResponse(403, { error: "invalid session secret" });
    }

    const pair = new WebSocketPair();
    const [client, server] = Object.values(pair);
    this.attach(side, server);
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

  private attach(side: Side, socket: WebSocket): void {
    const old = this.sockets.get(side);
    if (old) {
      old.close(1012, "replaced");
    }

    socket.accept();
    this.sockets.set(side, socket);
    this.flush(side);

    socket.addEventListener("message", (event) => {
      this.forward(side, event.data).catch(() => {
        socket.close(1003, "unsupported message");
      });
    });

    socket.addEventListener("close", (event) => this.detach(side, event.code, event.reason));
    socket.addEventListener("error", () => this.detach(side, 1011, "socket error"));
  }

  private detach(side: Side, code: number, reason: string): void {
    const socket = this.sockets.get(side);
    this.sockets.delete(side);
    const finalClose = code === FINAL_CLOSE_CODE && reason === FINAL_CLOSE_REASON;
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.close(1000, "closed");
    }
    if (!finalClose) {
      return;
    }
    const peer = this.sockets.get(peerSide(side));
    if (peer && peer.readyState === WebSocket.OPEN) {
      peer.close(FINAL_CLOSE_CODE, `${side} disconnected`);
    }
  }

  private async forward(from: Side, data: unknown): Promise<void> {
    const payload = await normalizePayload(data);
    this.deliver(peerSide(from), payload);
  }

  private deliver(to: Side, payload: RelayPayload): void {
    const socket = this.sockets.get(to);
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(payload);
      return;
    }
    const queue = this.pending.get(to) ?? [];
    queue.push(payload);
    while (queue.length > MAX_PENDING_MESSAGES) {
      queue.shift();
    }
    this.pending.set(to, queue);
  }

  private flush(side: Side): void {
    const socket = this.sockets.get(side);
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    const queue = this.pending.get(side) ?? [];
    while (queue.length > 0 && socket.readyState === WebSocket.OPEN) {
      const payload = queue.shift();
      if (payload !== undefined) {
        socket.send(payload);
      }
    }
  }

  private finalClose(side: Side): void {
    for (const current of [side, peerSide(side)] as Side[]) {
      const socket = this.sockets.get(current);
      this.sockets.delete(current);
      this.pending.set(current, []);
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close(FINAL_CLOSE_CODE, `${side} disconnected`);
      }
    }
  }
}

function peerSide(side: Side): Side {
  return side === "host" ? "client" : "host";
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
