# Cloudflare R2 Lifecycle

## Key Prefixes

Generated images are stored under lifecycle-aware prefixes:

- Free or anonymous images: `generations/free/{YYYY-MM}/{owner}-{generationID}.png`
- Paid images: `generations/paid/{YYYY-MM}/{owner}-{generationID}.png`

When an administrator tops up a user, existing objects for that user under `generations/free/` are promoted to `generations/paid/` and the database `r2_key` is updated. With R2 configured, the service copies the object to the paid key and deletes the old free key. Without R2 configured, local tests still verify the database key migration path.

## Cloudflare Rule

Create one lifecycle rule in the R2 bucket:

- Rule name: `expire-free-generations`
- Prefix: `generations/free/`
- Action: expire/delete objects
- Age: 7 days after object creation

Do not configure automatic expiration for `generations/paid/`.

## Validation

1. Generate an image before top-up and confirm its key starts with `generations/free/`.
2. Top up the user from the admin panel.
3. Confirm the old key is copied to `generations/paid/`, the old `generations/free/` object is deleted, and the `generations.r2_key` value is updated.
4. Generate another image after top-up and confirm its key starts with `generations/paid/`.
