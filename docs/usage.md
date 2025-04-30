# Using ncurl

ncurl allows you to describe HTTP requests in plain English, and it will generate and execute the appropriate request.

## Basic Usage

```bash
ncurl "get the latest posts from jsonplaceholder"
```

This will:
1. Parse your natural language request
2. Generate a proper HTTP request (in this case, to JSONPlaceholder's posts endpoint)
3. Execute the request
4. Display the response

## Command Options

| Option | Description |
|--------|-------------|
| `-t <seconds>` | Set timeout in seconds (default: 30) |
| `-m <model>` | Specify Anthropic model to use (default: claude-3-7-sonnet) |
| `-j` | Output response body as JSON only |
| `-v` | Verbose output (include request details) |
| `-version` | Show version information |
| `-debug` | Enable debug logging |

## Working with Command History

ncurl keeps track of your commands, allowing you to reference or rerun previous requests.

### Viewing History

```bash
ncurl -history
```

This displays your command history with numbers, status indicators, and timestamps.

### Searching History

```bash
ncurl -search "weather"
```

This will search your history for commands containing "weather".

### Rerunning Commands

```bash
ncurl -rerun 3
```

This reruns the 3rd command in your history.

### Interactive History Selection

```bash
ncurl -i
```

This launches interactive mode where you can browse and select a command from your history.

## Examples

Get data from a REST API:
```bash
ncurl "get user data for user id 5 from jsonplaceholder"
```

Post JSON data:
```bash
ncurl "post a new user with name 'John Doe' and email 'john@example.com' to jsonplaceholder"
```

Make an authenticated request:
```bash
ncurl "get my github profile using bearer token ghp_abc123"
```

Get weather data:
```bash
ncurl "get the current weather for New York City from the OpenWeatherMap API"
```