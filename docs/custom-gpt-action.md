# Private Custom GPT Recipe Capture

Use a private Custom GPT Action as the conversational client for the same capture API. This replaces the older ChatGPT plugin approach for this narrow personal workflow.

## Configure the GPT

1. Create a private GPT in ChatGPT.
2. Add instructions:

   > When I ask to save or import a recipe, call `captureRecipe`. Send the recipe's public HTTPS URL when available. If I pasted the recipe itself, send it as `sourceText`. Set `client` to `chatgpt-action`. Never claim the recipe was published; report the returned GitHub issue URL and explain that a draft PR will be created for review.

3. Add an Action by importing:

   `https://YOUR_RECIPE_BOX_DOMAIN/api/openapi.json`

4. Choose API-key authentication with bearer auth.
5. Use the same `RECIPE_CAPTURE_TOKEN` configured in Cloudflare, or provision a separate accepted token when multi-token support is added.
6. Test both a URL capture and pasted-text capture in Preview.

The GPT should remain private because its action can create import jobs. The Action only queues work; it cannot publish recipes or write repository contents.

Build a full ChatGPT App/MCP server only if the desired experience expands to searching the vault, displaying interactive recipe cards, editing pending imports, or managing review state inside ChatGPT.
