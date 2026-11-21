# Run PND

Install Docker with Compose v2. On Windows, select Linux containers in Docker Desktop.

1. Extract this deployment bundle into a new folder.
2. Copy `.env.example` to `.env`. Set the browser address (`PND_PUBLIC_URL`), generated hostname domain (`PND_DOMAIN`), and prefix (`PND_PREFIX`). The defaults support a local trial.
3. Open a terminal in that folder and run:

```sh
docker compose up -d
```

Open `http://localhost:8080` or your configured browser address. Initialization generates the encryption key once in the `pnd_keys` Docker volume. The application reads it through a read-only mount. Back up that volume; existing names require the same key. Keep the Compose project name stable and preserve volumes during upgrades.

The default service is public to anyone who can reach it. See [deployment settings](docs/DEPLOYMENT.md) for OIDC sign-in and geographic services.

To update, change `PND_IMAGE` in `.env` to the chosen release tag and run:

```sh
docker compose pull
docker compose up -d
```

Optional IP lookup downloads a separately licensed database:

```sh
docker compose -f compose.yaml -f compose.geoip-auto.yaml up -d
```

Use the same overlay files when upgrading a deployment that enables optional services. `docker compose down` preserves data volumes; `docker compose down -v` deletes them.

Image tags support exact versions (`26.2.3`), a release series (`26.2`), a year (`26`), and `latest`. Series/year/latest tags follow stable releases in their scope. Candidates use only their full version tag. The bundle defaults to an exact version.
