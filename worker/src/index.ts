/**
* Cert Service Cloudflare Worker
*/

export interface Env {
  // where the manifest files are stored
  CERT_MANIFESTS: KVNamespace;

  // simple security
  CERT_SERVICE_SALT: string;
}

const textEncoder = new TextEncoder();

function toHex(arrayBuffer: ArrayBuffer) {
  const byteArray = new Uint8Array(arrayBuffer);
  let result = "";
  for (const byte of byteArray) {
    result += byte.toString(16).padStart(2, "0");
  }
  return result;
}

async function validateChecksum(id: string, timestamp: string, checksum: string, salt: string): Promise<boolean> {
  const keyMaterial = textEncoder.encode(`${id}${timestamp}${salt}`);
  const digest = await crypto.subtle.digest("MD5", keyMaterial);
  const digestHex = toHex(digest);
  return digestHex === checksum.toLowerCase();
}

function makeManifestKey(timestamp: string) {
  if (!/^-?\d+$/.test(timestamp)) {
    throw new Error("invalid timestamp");
  }
  const value = Number(timestamp);
  if (!Number.isFinite(value)) {
    throw new Error("invalid timestamp");
  }

  const date = new Date(value * 1000);
  if (Number.isNaN(date.valueOf())) {
    throw new Error("invalid timestamp");
  }

  const year = date.getUTCFullYear();
  const month = `${date.getUTCMonth() + 1}`.padStart(2, "0");
  const day = `${date.getUTCDate()}`.padStart(2, "0");
  return `certs-${year}${month}${day}.json`;
}

async function fetchManifest(env: Env, key: string) {
  if (!env.CERT_MANIFESTS || typeof env.CERT_MANIFESTS.get !== "function") {
    throw new Error("CERT_MANIFESTS binding is not configured");
  }

  const manifest = await env.CERT_MANIFESTS.get(key, "arrayBuffer");
  if (!manifest) {
    return null;
  }

  return manifest;
}

function gzipResponse(data: string | ArrayBuffer | Blob | ArrayBufferView<ArrayBufferLike>) {
  const stream = new Blob([data]).stream().pipeThrough(new CompressionStream("gzip"));
  return new Response(stream, {
    status: 200,
    headers: {
      "Content-Type": "application/octet-stream",
      "Content-Encoding": "gzip",
      "Cache-Control": "no-store",
    },
  });
}


async function handleRequest(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const id = url.searchParams.get("a");
  const timestamp = url.searchParams.get("b");
  const checksum = url.searchParams.get("c");

  if (!id || !timestamp || !checksum) {
    return new Response("missing query parameters", { status: 400 });
  }

  if (!env.CERT_SERVICE_SALT) {
    return new Response("CERT_SERVICE_SALT binding is not configured", { status: 500 });
  }

  let checksumIsValid = false;
  try {
    checksumIsValid = await validateChecksum(id, timestamp, checksum, env.CERT_SERVICE_SALT);
  } catch (error) {
    console.error("failed to validate checksum", error);
    return new Response(null, { status: 500 });
  }

  if (!checksumIsValid) {
    return new Response(null, { status: 403 });
  }

  let manifestKey;
  try {
    manifestKey = makeManifestKey(timestamp);
  } catch (error) {
    console.error("failed to derive manifest key", error);
    return new Response(null, { status: 404 });
  }

  let manifest;
  try {
    manifest = await fetchManifest(env, manifestKey);
  } catch (error) {
    console.error("failed to fetch manifest", error);
    return new Response(null, { status: 500 });
  }

  if (!manifest) {
    return new Response(null, { status: 404 });
  }

  return gzipResponse(manifest);
}

export default {
  async fetch(request, env, _context): Promise<Response> {
    try {
      return await handleRequest(request, env);
    } catch (error) {
      console.error("unexpected error", error);
      return new Response(null, { status: 500 });
    }
  },
} satisfies ExportedHandler<Env>;
