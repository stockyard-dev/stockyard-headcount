# Stockyard Headcount

**Team directory and org chart — names, roles, contact info, who reports to whom**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted developer tools.

## Quick Start

```bash
docker run -p 9160:9160 -v headcount_data:/data ghcr.io/stockyard-dev/stockyard-headcount
```

Or with docker-compose:

```bash
docker-compose up -d
```

Open `http://localhost:9160` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9160` | HTTP port |
| `DATA_DIR` | `./data` | SQLite database directory |
| `HEADCOUNT_LICENSE_KEY` | *(empty)* | Pro license key |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 25 people | Unlimited people |
| Price | Free | $1.99/mo |

Get a Pro license at [stockyard.dev/tools/](https://stockyard.dev/tools/).

## Category

Operations & Teams

## License

Apache 2.0
