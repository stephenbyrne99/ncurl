# Troubleshooting ncurl

This guide helps you solve common issues when using ncurl.

## Common Issues

### API Key Problems

**Error**: `ANTHROPIC_API_KEY environment variable is required`

**Solution**: You need to set your Anthropic API key as an environment variable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

Make sure to replace `"your-api-key-here"` with your actual Anthropic API key.

### Request Timeouts

**Error**: `Request timed out after 30 seconds`

**Solution**: Some APIs may take longer to respond. Try increasing the timeout:

```bash
ncurl -t 60 "your request here"
```

This sets the timeout to 60 seconds.

### Unclear Requests

**Error**: The request generated doesn't match what you intended.

**Solution**:
1. Be more specific in your request
2. Use the `-v` flag to see the request details before execution
3. Try providing more context or specifying the exact API endpoint

For example, instead of:
```bash
ncurl "get weather"
```

Try:
```bash
ncurl "get current weather for New York from OpenWeatherMap API using API key xyz123"
```

### History File Issues

**Error**: `Failed to save command to history: permission denied`

**Solution**: Check the permissions on your `~/.ncurl` directory:

```bash
ls -la ~/.ncurl
chmod 755 ~/.ncurl
chmod 644 ~/.ncurl/history.json
```

## Getting Help

If you're still experiencing issues:

1. Run with debug mode to get more information:
   ```bash
   ncurl -debug "your request"
   ```

2. Check the GitHub repository for known issues or to file a bug report:
   [https://github.com/stephenbyrne99/ncurl/issues](https://github.com/stephenbyrne99/ncurl/issues)

3. Make sure you're using the latest version:
   ```bash
   ncurl -version
   ```

## API Limitations

Note that some APIs:
- Require authentication
- Have rate limits
- Have specific endpoint structures

For these cases, be as specific as possible in your request, including authentication details when necessary.