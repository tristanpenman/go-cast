# Worker

This directory contains a Cloudflare Worker implementation of the cert service.

This is based on the [Golang version](../cmd/cert-service/main.go) of the cert service.

## Wrangler

This project was initialised using Cloudflare's [Wrangler](https://developers.cloudflare.com/workers/wrangler/) tool.

## Prerequisites

Install dependencies and authenticate with Wrangler:

```bash
pnpm i
wrangler login
```

### Namespace

The `CERT_MANIFEST` KV namespace must be created:

```bash
npx wrangler kv namespace create CERT_MANIFESTS
```

You'll need to copy the returned namespace ID, and update your [wranger.jsonc](./wrangler.jsonc) file accordingly:

```json
  "kv_namespaces": [
    {
      "binding": "CERT_MANIFESTS",
      "id": "<YOUR ID HERE>",
      "remote": true
    }
  ]
```

### Manifests

Once this is created, individual cert manifests can be uploaded:

```bash
npx wrangler kv put \
  --binding=CERT_MANIFESTS certs-20260213.json \
  --path ../path/to/certs/20260213.json
```

You can repeat this for each day/version you need available.

### Salt

The service also depends on a salt, which is used to provide basic security:

```bash
npx wrangler secret put CERT_SERVICE_SALT
```

Set this to something [super secure](https://www.urbandictionary.com/define.php?term=hunter2).

### Deployment

```bash
npx wranger deploy
```
