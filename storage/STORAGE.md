# MinIO Object Storage — Setup

MinIO wordt gebruikt als S3-compatible object storage voor documentbijlagen.

## Lokale ontwikkelomgeving

### Starten met Docker Compose

```bash
docker compose up -d minio
```

Dit start MinIO op:
- **S3 API**: `http://localhost:9000`
- **Web console**: `http://localhost:9011` (login: `minioadmin` / `minioadmin`)

### Volledige stack (PostgreSQL + MinIO)

```bash
docker compose up -d
```

## Environment variabelen

| Variabele | Default | Beschrijving |
|---|---|---|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO S3 API endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | Toegangssleutel |
| `MINIO_SECRET_KEY` | `minioadmin` | Geheime sleutel |
| `MINIO_BUCKET` | `gesprekken-documenten` | Bucketnaam voor documenten |
| `MINIO_USE_SSL` | `false` | SSL gebruiken voor verbinding |

Deze staan in `.env` (en `.env.example`).

## Bucket initialisatie

De Go applicatie maakt de bucket **automatisch aan** bij het opstarten als deze nog niet bestaat. Er is geen apart init-script nodig.

## Bestandsopslag structuur

Bestanden worden opgeslagen met de key:
```
<brontype>/<documentId>.<extensie>
```

Bijvoorbeeld:
```
gespreksbijlage/a1b2c3d4-5678-90ab-cdef-1234567890ab.pdf
```

## Troubleshooting

| Probleem | Oplossing |
|---|---|
| `MinIO verbinding mislukt` | Controleer of MinIO draait: `docker compose ps` |
| Bucket access denied | Controleer `MINIO_ACCESS_KEY` en `MINIO_SECRET_KEY` in `.env` |
| Upload timeout | Controleer of `MINIO_ENDPOINT` bereikbaar is |
| Web console niet beschikbaar | Controleer of poort 9011 vrij is: `netstat -an \| findstr 9011` |

## Productie-overwegingen

- Verander standaard wachtwoorden (`MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`)
- Schakel SSL in (`MINIO_USE_SSL=true`) achter een reverse proxy
- Configureer bucket lifecycle policies voor cleanup
- Overweeg S3-compatible cloud storage (AWS S3, Azure Blob) als vervanging
