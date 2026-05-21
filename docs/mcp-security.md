# LiteDock MCP Security Notes

## Scope
LiteDock MCP exposes only `tools/list` and `tools/call`. The following MCP features are NOT implemented and should not be relied upon:
- `resources` - not supported
- `prompts` - not supported
- `sampling` - not supported

## Confirmation Bypass Prevention
All modifying container operations (stop, restart, delete) require a valid HMAC-SHA256 confirmation token regardless of invocation path (REST `/assistant/execute`, WebSocket stream, or MCP `tools/call`). The token is validated by the same `TokenValidator` service.

## Audit
All MCP calls are logged to the AI audit log with `source="mcp"`.
