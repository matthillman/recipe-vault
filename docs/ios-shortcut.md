# iPhone Share Sheet Shortcut

Create a Shortcut named **Save to Recipe Vault**.

## Shortcut Configuration

1. Open the shortcut details, enable **Show in Share Sheet**, and accept **URLs**, **Safari web pages**, and **Text**.
2. Add **Get URLs from Shortcut Input**.
3. Add an **If** action that checks whether the URL result has any value.
4. In the URL branch, create a Dictionary containing:
   - `url`: first URL
   - `client`: `ios-shortcut`
5. In the Otherwise branch, add **Get Text from Shortcut Input**, then create a Dictionary containing:
   - `sourceText`: that text
   - `client`: `ios-shortcut`
6. Add **Get Contents of URL** after the If block:
   - URL: `https://YOUR_RECIPE_BOX_DOMAIN/api/import`
   - Method: `POST`
   - Request Body: `JSON`, using the Dictionary
   - Header `Authorization`: `Bearer YOUR_RECIPE_CAPTURE_TOKEN`
   - Header `Content-Type`: `application/json`
7. Read `issueUrl` from the response Dictionary.
8. Show a notification saying **Recipe queued for review**.
9. Optionally add **Choose from Menu** with **Open issue** and **Done**; open `issueUrl` for the first choice.

The normal Safari flow is one share-sheet tap plus the Shortcut selection. The same URL branch supports public recipe pages, public PDF URLs, and public PNG, JPEG, WebP, or non-animated GIF URLs. A Safari PDF viewer URL therefore needs no Shortcut changes.

This URL-only media support does not upload a photo or file from the device. Sharing an item directly from Photos or Files requires a later binary-upload flow; for now, the image or PDF must already have a public HTTPS URL. Shared selected text remains the fallback for sites that block server-side extraction.

Treat the Shortcut as a secret because it contains the capture bearer token. Do not publish its iCloud sharing link publicly.
