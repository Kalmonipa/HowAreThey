# HowAreThey
A reminder system to keep in touch with your friends


### Docker Config
| Environment Variable | Details | Example | Default |
|---|---|---|---|
| DISCORD_WEBHOOK | Provide a Discord webhook to send notifications to Discord. Not providing a webhook will only log the events, it won't send the notification anywhere | `https://discord.com/api/webhooks/myexamplewebhook` | N/A |
| CRON | [Cron expression](https://crontab.guru/) to define how often a friend will get picked. By default, runs weekly. Use integer format for each field. | `0 0 7 * * 1` | `@weekly` |

### Example docker-compose.yaml
```
version: '3'

services:
  howarethey:
    container_name: howarethey
    image: kalmonipa/howarethey:v0.7
    environment:
        # Discord Webhook to send notifications too. Read the discord webhook docs to get one
      - DISCORD_WEBHOOK=https://discord.com/api/webhooks/myexamplewebhook
        # Picks a friend at 7am every Monday
      - CRON=0 0 7 * * 1
    ports:
      # The Web UI (enabled by --api.insecure=true)
      - "8022:8080"
    volumes:
        # Define where your persistent storage goes too
      - path/to/friend.db:/sql/friends.db
```
