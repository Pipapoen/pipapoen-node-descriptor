# Pipapoen Node Descriptor

[![AI Disclosure: AGAIN](https://img.shields.io/badge/AI%20Disclosure-AGAIN-8055CC?style=flat-square)](https://github.com/Doner357/AGAIN)
![Status: Release candidate](https://img.shields.io/badge/status-release%20candidate-64748B?style=flat-square)
[![License: MIT](https://img.shields.io/badge/license-MIT-2563EB?style=flat-square)](LICENSE)
![Platforms: Windows and Linux](https://img.shields.io/badge/platforms-Windows%20%7C%20Linux-475569?style=flat-square)
![Backend: Go](https://img.shields.io/badge/backend-Go-007D9C?style=flat-square)
![Frontend: Vite and TypeScript](https://img.shields.io/badge/frontend-Vite%20%2B%20TypeScript-3178C6?style=flat-square)
![Deployment: Docker Compose](https://img.shields.io/badge/deployment-Docker%20Compose-2496ED?style=flat-square)
[![CI](https://github.com/Pipapoen/pipapoen-node-descriptor/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/Pipapoen/pipapoen-node-descriptor/actions/workflows/ci.yml)

PND turns selected node information into a consistent hostname and lets you recover that information later. Choose the details to include, preview the available space, and copy the generated name.

The web interface includes English and Traditional Chinese, light/dark/system themes, location helpers, and data import/export.

[Try PND in your browser](https://pipapoen.github.io/pipapoen-node-descriptor/).

## Deploy with Docker Compose

Install Docker with Compose v2. On Windows, use Docker Desktop with Linux containers.

1. Download `pnd-<version>-compose.zip` from [Releases](https://github.com/Pipapoen/pipapoen-node-descriptor/releases) and extract it.
2. Copy `.env.example` to `.env`. Set your domain and browser address, or keep the defaults for a local trial.
3. Open a terminal in that folder and run:

```sh
docker compose up -d
```

Open `http://localhost:8080` or your configured address. PND generates its encryption key on first startup and preserves it in a Docker volume. Keep a backup of that volume to decode existing names.

The default service is public to anyone who can reach it. See [Deployment and configuration](docs/DEPLOYMENT.md) for SSO, location services, updates, and source builds.

## Create a hostname

1. Keep the generated Node ID, or **Import data** for an existing node.
2. Fill in the fields you want to include. Leave optional fields empty. Provider suggestions help you choose resource combinations.
3. Set the prefix and domain, choose **Compact** or **Fixed**, then select **Generate hostname**.
4. **Copy** the result and **Export data** to retain the node's identity and selections. Create DNS records through your DNS provider.

![Encoding node information and generating a hostname](docs/images/encode.png)

## Find a location

Open **Find a location** below GeoNames ID. Open the country dropdown and type **United States** into its search box. Then open the city dropdown, search for **New York**, select the city, and press **Use this location**.

The country/city list covers major cities and capitals. Place-name and IP queries are also available when configured.

![Country and city filtering beneath GeoNames ID](docs/images/location.png)

## Decode or update a node

Open **Decode**, paste a hostname, label, or URL, and select **Decode hostname**. Location results show the GeoNames ID together with available country and city names. Use **Edit this node** to update its fields while retaining the Node ID, or **Export data** to save the recovered information.

![Decoding a hostname and recovering node details](docs/images/decode.png)

Screenshots use example data. Keep exported node files when upgrading the MVP.

## Documentation

- [Deployment and configuration](docs/DEPLOYMENT.md)
- [Technical specification](docs/TECHNICAL.md)
- [Contributing](CONTRIBUTING.md)
- [Release operations](docs/RELEASING.md)
- [Translation contributions](docs/TRANSLATING.md)
- [Third-party notices](THIRD_PARTY_NOTICES.md)
